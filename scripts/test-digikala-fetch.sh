#!/bin/bash

# Test script to fetch real price from Digikala
# This shows how the scraper should work

DKP_ID="20758981"

echo "🔍 Fetching price for product: dkp-$DKP_ID"
echo ""

# Try Digikala's API endpoint
echo "=== Method 1: Digikala API ==="
RESPONSE=$(curl -s "https://api.digikala.com/v1/product/dkp-$DKP_ID/" \
  -H "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36" \
  -H "Accept: application/json")

# Check if we got data
if echo "$RESPONSE" | jq -e '.data.product' > /dev/null 2>&1; then
    echo "✅ API Response received!"
    
    # Extract price
    PRICE=$(echo "$RESPONSE" | jq -r '.data.product.default_variant.price.selling_price // .data.product.default_variant.price.rrp_price // "N/A"')
    TITLE=$(echo "$RESPONSE" | jq -r '.data.product.title_fa // "N/A"')
    SELLER=$(echo "$RESPONSE" | jq -r '.data.product.default_variant.seller.title // "digikala"')
    
    echo ""
    echo "📦 Product: $TITLE"
    echo "💰 Price: $(echo $PRICE | sed 's/\(.\)\(.\)\(.\)$/,\1\2\3/') Toman"
    echo "🏪 Seller: $SELLER"
    echo ""
    
    # Show the JSON structure
    echo "=== Full Price Data ==="
    echo "$RESPONSE" | jq '.data.product.default_variant.price'
else
    echo "❌ API request failed or returned no data"
    echo ""
    echo "Response preview:"
    echo "$RESPONSE" | head -c 500
fi

echo ""
echo "=== Method 2: Web Scraping (HTML) ==="
echo "Fetching HTML page..."

HTML=$(curl -s "https://www.digikala.com/product/dkp-$DKP_ID/" \
  -H "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

# Look for Next.js data or embedded JSON
if echo "$HTML" | grep -q "__NEXT_DATA__"; then
    echo "✅ Found __NEXT_DATA__ in HTML"
    echo "$HTML" | grep -o '__NEXT_DATA__[^<]*' | sed 's/__NEXT_DATA__.*=//' | jq -r '.props.pageProps.initialState.product.product.default_variant.price.selling_price // "Could not extract price"' 2>/dev/null || echo "Price extraction needs adjustment"
else
    echo "❌ __NEXT_DATA__ not found, page might be client-side rendered"
fi

echo ""
echo "💡 Recommendation: Use Digikala's API endpoint for reliable data extraction"
