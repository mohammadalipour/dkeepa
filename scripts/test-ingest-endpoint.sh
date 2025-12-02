#!/bin/bash
# Test the product ingest endpoint

set -e

echo "🧪 Testing Product Ingest Endpoint"
echo "=================================="
echo ""

API_URL="http://localhost:8080/api/v1/products/ingest"

# Test data (from the Lenovo laptop we scraped)
TEST_DATA='{
  "dkp_id": "20758981",
  "variant_id": "73675351",
  "title": "لپ تاپ 15.6 اینچی لنوو مدل IdeaPad Slim 3",
  "price": 306990000,
  "rrp_price": 325000000,
  "seller_name": "دیجی‌کالا",
  "is_active": true,
  "rch_token": "89e796febf77"
}'

echo "📤 Sending test data to: $API_URL"
echo ""
echo "Request body:"
echo "$TEST_DATA" | jq .
echo ""

# Send POST request
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d "$TEST_DATA")

# Extract status code and body
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "📥 Response (HTTP $HTTP_CODE):"
echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
echo ""

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Success! Product data ingested"
    echo ""
    echo "🔍 Verify in database:"
    echo "  docker exec keepa-timescaledb psql -U postgres -d keepa -c \"SELECT * FROM products WHERE dkp_id = '20758981';\""
    echo "  docker exec keepa-timescaledb psql -U postgres -d keepa -c \"SELECT time, price, seller_id FROM price_history WHERE dkp_id = '20758981' ORDER BY time DESC LIMIT 5;\""
else
    echo "❌ Failed! HTTP $HTTP_CODE"
    exit 1
fi
