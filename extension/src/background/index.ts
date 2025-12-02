// Background service worker
console.log('Keepa background service worker started');

// Listen for extension installation
chrome.runtime.onInstalled.addListener(() => {
    console.log('Keepa extension installed');

    // Set up periodic alarm for checking hot products
    chrome.alarms.create('checkHotProducts', {
        periodInMinutes: 60 // Check every hour
    });
});

// Handle alarm events
chrome.alarms.onAlarm.addListener((alarm) => {
    if (alarm.name === 'checkHotProducts') {
        console.log('Checking hot products...');
        // TODO: Implement hot products check logic
    }
});

// Handle messages from content scripts or popup
chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
    console.log('Message received:', message);

    if (message.type === 'GET_PRICE_HISTORY') {
        fetchPriceHistory(message.dkpId, message.variantId)
            .then(data => sendResponse({ success: true, data }))
            .catch(error => sendResponse({ success: false, error: error.message }));
        return true; // Keep channel open for async response
    }

    if (message.type === 'SEND_PRODUCT_DATA') {
        sendProductDataToBackend(message.data)
            .then(() => sendResponse({ success: true }))
            .catch(error => sendResponse({ success: false, error: error.message }));
        return true; // Keep channel open for async response
    }
});

// Fetch price history from backend API
async function fetchPriceHistory(dkpId: string, variantId?: string | null) {
    const API_URL = 'http://localhost:8080';
    let url = `${API_URL}/api/v1/products/${dkpId}/history`;

    if (variantId) {
        url += `?variant_id=${variantId}`;
    }

    const response = await fetch(url);

    if (!response.ok) {
        throw new Error(`API request failed: ${response.statusText}`);
    }

    return response.json();
}

// Send scraped product data to backend
async function sendProductDataToBackend(productData: any) {
    const API_URL = 'http://localhost:8080';
    const url = `${API_URL}/api/v1/products/ingest`;

    console.log('Sending product data to backend:', productData);

    const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            dkp_id: productData.dkpId,
            variant_id: productData.variantId,
            title: productData.title,
            price: productData.price,
            rrp_price: productData.rrpPrice,
            seller_name: productData.sellerName,
            is_active: productData.isActive,
            rch_token: productData.rchToken,
        }),
    });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to send data to backend: ${response.status} ${errorText}`);
    }

    return response.json();
}

export { };
