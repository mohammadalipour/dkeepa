import ReactDOM from 'react-dom/client';
import PriceChart from './PriceChart';
import ShadowRoot from './ShadowRoot';

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

// Initialize the extension
function init() {
    const dkpId = getProductId();

    if (!dkpId) {
        console.log('No product ID found in URL');
        return;
    }

    const variantId = getVariantId();
    console.log('Product ID:', dkpId, 'Variant ID:', variantId);

    // Wait a bit for the page to load, then inject price chart
    setTimeout(() => {
        injectPriceChart(dkpId, variantId);
    }, 1000);
}

// Wait for DOM to be ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}
