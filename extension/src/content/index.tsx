import ReactDOM from 'react-dom/client';
import PriceChart from './PriceChart';
import ShadowRoot from './ShadowRoot';
import { scrapeProductData } from './digikalaScraper';

console.log('Keepa content script loaded');

// Extract product ID from URL
function getProductId(): string | null {
    const url = window.location.href;
    const match = url.match(/dkp-(\d+)/);
    return match ? match[1] : null;
}

// Extract variant ID from URL
function getVariantId(): string | null {
    const url = window.location.href;
    const match = url.match(/variant_id=(\d+)/);
    return match ? match[1] : null;
}


// Create shadow DOM host and inject our component
function injectPriceChart(dkpId: string, variantId: string | null) {
    // Create host element
    const host = document.createElement('div');
    host.id = 'keepa-price-chart-host';
    host.style.cssText = `
        margin: 20px auto;
        max-width: 1200px;
        padding: 20px;
        background: white;
        border-radius: 8px;
        box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        z-index: 1000;
    `;

    // Try to find main content area, otherwise use body
    const mainContent = document.querySelector('main') ||
        document.querySelector('[role="main"]') ||
        document.querySelector('#app') ||
        document.body;

    // Insert at the beginning of main content
    mainContent.insertBefore(host, mainContent.firstChild);

    // Create Shadow DOM
    const shadowRoot = host.attachShadow({ mode: 'open' });

    // Create root for React
    const root = ReactDOM.createRoot(shadowRoot);

    root.render(
        <ShadowRoot>
            <PriceChart dkpId={dkpId} variantId={variantId} />
        </ShadowRoot>
    );
}

// Scrape and send product data to backend
async function scrapeAndSendData(dkpId: string, variantId: string | null) {
    try {
        console.log('Scraping product data...');
        const productData = await scrapeProductData(dkpId, variantId);

        if (productData) {
            console.log('Product data scraped:', productData);
            
            // Send to background script to forward to backend
            chrome.runtime.sendMessage({
                type: 'SEND_PRODUCT_DATA',
                data: productData
            }, (response) => {
                if (response?.success) {
                    console.log('✅ Product data sent to backend successfully');
                } else {
                    console.warn('⚠️ Failed to send to backend:', response?.error);
                }
            });
        } else {
            console.warn('Could not scrape product data');
        }
    } catch (error) {
        console.error('Error scraping data:', error);
    }
}

// Initialize the extension
function init() {
    const dkpId = getProductId();

    if (!dkpId) {
        console.log('No product ID found in URL');
        return;
    }

    const variantId = getVariantId();
    console.log('Product ID:', dkpId, 'Variant ID:', variantId);

    // Wait a bit for the page to load, then scrape and send data
    setTimeout(() => {
        injectPriceChart(dkpId, variantId);
        
        // Scrape and send product data to backend (wait a bit more for page to fully load)
        setTimeout(() => {
            scrapeAndSendData(dkpId, variantId);
        }, 2000);
    }, 1000);
}

// Wait for DOM to be ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}
