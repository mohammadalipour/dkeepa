#!/bin/bash

# Keepa - Check Product Data
# Usage: ./scripts/check-product.sh <dkp_id>

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <dkp_id>"
    echo "Example: $0 20758981"
    exit 1
fi

DKP_ID=$1

echo "📊 Checking data for product: $DKP_ID"
echo ""

echo "=== Product Info ==="
docker exec keepa-timescaledb psql -U postgres -d keepa -c \
    "SELECT * FROM products WHERE dkp_id = '$DKP_ID';"

echo ""
echo "=== Price History (Last 10 entries) ==="
docker exec keepa-timescaledb psql -U postgres -d keepa -c \
    "SELECT time, price, seller_id, is_buy_box FROM price_history WHERE dkp_id = '$DKP_ID' ORDER BY time DESC LIMIT 10;"

echo ""
echo "=== Statistics ==="
docker exec keepa-timescaledb psql -U postgres -d keepa -c \
    "SELECT 
        COUNT(*) as total_records,
        MIN(price) as min_price,
        MAX(price) as max_price,
        AVG(price)::INTEGER as avg_price
    FROM price_history 
    WHERE dkp_id = '$DKP_ID';"
