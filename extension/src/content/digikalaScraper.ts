/**
 * Digikala Scraper - Extracts product data from Digikala pages
 */

export interface ProductData {
    dkpId: string;
    variantId: string | null;
    title: string;
    price: number;
    rrpPrice: number;
    sellerName: string;
    isActive: boolean;
    rchToken: string | null;
}

/**
 * Extract the _rch parameter from the page
 * This token is used by Digikala API for anti-bot protection
 */
export function extractRchToken(): string | null {
    try {
        // Method 1: Check URL parameters (sometimes present)
        const urlParams = new URLSearchParams(window.location.search);
        const rchFromUrl = urlParams.get('_rch');
        if (rchFromUrl) {
            console.log('Found _rch in URL:', rchFromUrl);
            return rchFromUrl;
        }

        // Method 2: Extract from Next.js data script
        const nextDataScript = document.getElementById('__NEXT_DATA__');
        if (nextDataScript) {
            try {
                const nextData = JSON.parse(nextDataScript.textContent || '{}');
                // Check if _rch is in the buildId or query
                if (nextData.query?._rch) {
                    console.log('Found _rch in Next.js data:', nextData.query._rch);
                    return nextData.query._rch;
                }
                if (nextData.buildId) {
                    // Sometimes _rch is encoded in buildId
                    console.log('Next.js buildId:', nextData.buildId);
                }
            } catch (e) {
                console.warn('Failed to parse Next.js data:', e);
            }
        }

        // Method 3: Look for it in inline scripts or window object
        const scripts = document.querySelectorAll('script');
        for (const script of scripts) {
            const content = script.textContent || '';
            const rchMatch = content.match(/_rch['":\s]+['"]?([a-f0-9]{12,})['"]?/i);
            if (rchMatch) {
                console.log('Found _rch in script:', rchMatch[1]);
                return rchMatch[1];
            }
        }

        // Method 4: Check window object
        if ((window as any).__NEXT_DATA__?.query?._rch) {
            return (window as any).__NEXT_DATA__.query._rch;
        }

        console.warn('Could not find _rch token on page');
        return null;
    } catch (error) {
        console.error('Error extracting _rch token:', error);
        return null;
    }
}

/**
 * Intercept and extract _rch from API requests
 * This uses the Fetch API interception to capture the token when the page makes requests
 */
export function interceptRchToken(callback: (token: string) => void) {
    const originalFetch = window.fetch;
    window.fetch = function (...args: Parameters<typeof fetch>) {
        let url = '';
        if (typeof args[0] === 'string') {
            url = args[0];
        } else if (args[0] instanceof Request) {
            url = args[0].url;
        } else if (args[0] instanceof URL) {
            url = args[0].toString();
        }
        
        // Check if this is a Digikala API request with _rch
        if (url.includes('api.digikala.com') && url.includes('_rch=')) {
            const match = url.match(/_rch=([a-f0-9]{12,})/i);
            if (match) {
                console.log('Intercepted _rch from API request:', match[1]);
                callback(match[1]);
            }
        }
        
        return originalFetch.apply(this, args);
    };
}

/**
 * Scrape product data from the current Digikala page
 */
export async function scrapeProductData(dkpId: string, variantId: string | null): Promise<ProductData | null> {
    try {
        console.log('Scraping product data for:', dkpId, 'variant:', variantId);

        // First try to get _rch token
        let rchToken = extractRchToken();

        // If we couldn't find it, set up interception and wait briefly
        if (!rchToken) {
            console.log('Waiting for _rch token from API requests...');
            await new Promise<void>((resolve) => {
                let found = false;
                interceptRchToken((token) => {
                    if (!found) {
                        rchToken = token;
                        found = true;
                        resolve();
                    }
                });
                // Timeout after 3 seconds
                setTimeout(() => resolve(), 3000);
            });
        }

        // Method 1: Try to fetch from Digikala API with _rch token
        if (rchToken) {
            const apiData = await fetchFromDigikalaApi(dkpId, variantId, rchToken);
            if (apiData) {
                return apiData;
            }
        }

        // Method 2: Fallback to extracting from page DOM
        console.log('Falling back to DOM extraction');
        return extractFromPageDOM(dkpId, variantId, rchToken);

    } catch (error) {
        console.error('Error scraping product data:', error);
        return null;
    }
}

/**
 * Fetch product data from Digikala API using _rch token
 */
async function fetchFromDigikalaApi(
    dkpId: string,
    variantId: string | null,
    rchToken: string
): Promise<ProductData | null> {
    try {
        let apiUrl = `https://api.digikala.com/v2/product/${dkpId}/?_rch=${rchToken}`;
        if (variantId) {
            apiUrl += `&variant_id=${variantId}`;
        }

        console.log('Fetching from Digikala API:', apiUrl);
        const response = await fetch(apiUrl);

        if (!response.ok) {
            console.warn('API request failed:', response.status);
            return null;
        }

        const data = await response.json();

        if (data.status !== 200 || !data.data?.product) {
            console.warn('Invalid API response:', data);
            return null;
        }

        const product = data.data.product;
        const defaultVariant = product.default_variant;

        return {
            dkpId,
            variantId: variantId || defaultVariant.id?.toString() || null,
            title: product.title_fa || product.title_en || '',
            price: defaultVariant.price?.selling_price || defaultVariant.price?.rrp_price || 0,
            rrpPrice: defaultVariant.price?.rrp_price || 0,
            sellerName: defaultVariant.seller?.title || 'دیجی‌کالا',
            isActive: product.status === 'marketable',
            rchToken,
        };
    } catch (error) {
        console.error('Error fetching from Digikala API:', error);
        return null;
    }
}

/**
 * Extract product data from page DOM (fallback method)
 */
function extractFromPageDOM(
    dkpId: string,
    variantId: string | null,
    rchToken: string | null
): ProductData | null {
    try {
        // Try JSON-LD first
        const jsonLd = document.querySelector('script[type="application/ld+json"]');
        if (jsonLd) {
            try {
                const data = JSON.parse(jsonLd.textContent || '{}');
                if (data['@type'] === 'Product') {
                    return {
                        dkpId,
                        variantId,
                        title: data.name || '',
                        price: parseFloat(data.offers?.price || '0') * 10, // Convert to Rials
                        rrpPrice: parseFloat(data.offers?.price || '0') * 10,
                        sellerName: data.offers?.seller?.name || 'دیجی‌کالا',
                        isActive: data.offers?.availability === 'https://schema.org/InStock',
                        rchToken,
                    };
                }
            } catch (e) {
                console.warn('Failed to parse JSON-LD:', e);
            }
        }

        // Try extracting from page title and price elements
        const titleElement = document.querySelector('h1[data-title-en], h1.text-h4, h1');
        const priceElement = document.querySelector('[data-selling-price], .text-h4.text-neutral-900');

        if (titleElement && priceElement) {
            const priceText = priceElement.textContent?.replace(/[^\d]/g, '') || '0';
            return {
                dkpId,
                variantId,
                title: titleElement.textContent?.trim() || '',
                price: parseInt(priceText) * 10, // Convert to Rials
                rrpPrice: parseInt(priceText) * 10,
                sellerName: 'دیجی‌کالا',
                isActive: true,
                rchToken,
            };
        }

        console.warn('Could not extract product data from DOM');
        return null;
    } catch (error) {
        console.error('Error extracting from DOM:', error);
        return null;
    }
}
