#!/bin/bash

# Insert test price data for a product
# Usage: ./scripts/insert-test-data.sh <dkp_id>

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <dkp_id>"
    echo "Example: $0 20758981"
    exit 1
fi

DKP_ID=$1

echo "📊 Inserting test data for product: $DKP_ID"

# Insert product
docker exec keepa-timescaledb psql -U postgres -d keepa -c "
INSERT INTO products (dkp_id, title, is_active, last_scraped_at)
VALUES ('$DKP_ID', 'Test Product - Lenovo IdeaPad', true, NOW())
ON CONFLICT (dkp_id) DO UPDATE SET last_scraped_at = NOW();
"

# Insert price history (last 30 days with some variation)
docker exec keepa-timescaledb psql -U postgres -d keepa -c "
INSERT INTO price_history (time, dkp_id, price, seller_id, is_buy_box)
SELECT 
    NOW() - (interval '1 day' * generate_series),
    '$DKP_ID',
    45000000 + (random() * 5000000)::INTEGER,
    'digikala',
    true
FROM generate_series(0, 29);
"

echo "✅ Test data inserted successfully!"
echo ""
echo "📊 Check the data:"
echo "   ./scripts/check-product.sh $DKP_ID"
echo ""
echo "🌐 Now refresh the Digikala page to see the chart!"
