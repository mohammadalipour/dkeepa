#!/bin/bash

# Complete Price Tracking Pipeline
# This script runs the full workflow to fetch all products, variants, and prices

set -e  # Exit on error

echo "🚀 Starting Complete Price Tracking Pipeline for Mobile Phones"
echo "================================================================"

# Configuration
CATEGORY="mobile-phone"
DB_HOST="${DB_HOST:-localhost}"
MAX_PRODUCTS="${MAX_PRODUCTS:-0}"  # 0 = unlimited
CONCURRENCY="${CONCURRENCY:-3}"
DELAY="${DELAY:-1500}"

echo "📋 Configuration:"
echo "   Category: $CATEGORY"
echo "   Max Products: $MAX_PRODUCTS (0 = unlimited)"
echo "   Concurrency: $CONCURRENCY workers"
echo "   Delay: ${DELAY}ms"
echo ""

# Step 1: Fetch ALL Products
echo "================================"
echo "📱 STEP 1: Fetching ALL Products"
echo "================================"
echo "This will fetch all mobile phone products from Digikala..."
echo "Expected: ~3,680 products, ~45-50 minutes with 3 workers"
echo ""

START_TIME=$(date +%s)

DB_HOST=$DB_HOST ./category-crawler-v2 \
    --category=$CATEGORY \
    --max=$MAX_PRODUCTS \
    --concurrency=$CONCURRENCY \
    --delay=$DELAY

STEP1_END=$(date +%s)
STEP1_DURATION=$((STEP1_END - START_TIME))
echo "✅ Step 1 completed in $(($STEP1_DURATION / 60)) minutes"
echo ""

# Check products count
PRODUCT_COUNT=$(docker exec keepa-timescaledb psql -U postgres -d keepa -t -c "SELECT COUNT(*) FROM products WHERE category='$CATEGORY'")
echo "📊 Total products in database: $PRODUCT_COUNT"
echo ""

# Step 2: Fetch ALL Variants
echo "===================================="
echo "🔍 STEP 2: Fetching ALL Variants"
echo "===================================="
echo "This will discover all variants for each product..."
echo "Expected time: ~3-6 hours for 3,680 products (2s delay per product)"
echo ""

DB_HOST=$DB_HOST ./variant-crawler

STEP2_END=$(date +%s)
STEP2_DURATION=$((STEP2_END - STEP1_END))
echo "✅ Step 2 completed in $(($STEP2_DURATION / 60)) minutes"
echo ""

# Check variants count
VARIANT_COUNT=$(docker exec keepa-timescaledb psql -U postgres -d keepa -t -c "SELECT COUNT(*) FROM product_variants WHERE is_active=true")
echo "📊 Total active variants in database: $VARIANT_COUNT"
echo ""

# Step 3: Track Initial Prices
echo "=================================="
echo "💰 STEP 3: Tracking Initial Prices"
echo "=================================="
echo "This will fetch current prices for all variants..."
echo "Expected time: ~2-4 hours for all variants"
echo ""

DB_HOST=$DB_HOST ./price-tracker

STEP3_END=$(date +%s)
STEP3_DURATION=$((STEP3_END - STEP2_END))
echo "✅ Step 3 completed in $(($STEP3_DURATION / 60)) minutes"
echo ""

# Check price logs count
PRICE_COUNT=$(docker exec keepa-timescaledb psql -U postgres -d keepa -t -c "SELECT COUNT(*) FROM price_history WHERE dkp_id IN (SELECT dkp_id FROM products WHERE category='$CATEGORY')")
echo "📊 Total price logs in database: $PRICE_COUNT"
echo ""

# Final Statistics
TOTAL_DURATION=$((STEP3_END - START_TIME))
echo "================================================================"
echo "🎉 PIPELINE COMPLETED!"
echo "================================================================"
echo "⏱️  Total Duration: $(($TOTAL_DURATION / 3600))h $(($TOTAL_DURATION % 3600 / 60))m"
echo "📦 Products: $PRODUCT_COUNT"
echo "🔍 Variants: $VARIANT_COUNT"
echo "💰 Price Logs: $PRICE_COUNT"
echo ""
echo "📊 Breakdown:"
echo "   Step 1 (Products):  $(($STEP1_DURATION / 60))m"
echo "   Step 2 (Variants):  $(($STEP2_DURATION / 60))m"
echo "   Step 3 (Prices):    $(($STEP3_DURATION / 60))m"
echo ""
echo "✅ All mobile phone products, variants, and prices are now in the database!"
echo "================================================================"

# Show sample data
echo ""
echo "📋 Sample Data (Top 5 Products with Variants and Prices):"
docker exec keepa-timescaledb psql -U postgres -d keepa -c "
SELECT 
    p.dkp_id,
    LEFT(p.title, 40) as title,
    COUNT(DISTINCT pv.variant_id) as variants,
    COUNT(ph.time) as price_logs,
    MAX(ph.price) as latest_price
FROM products p
LEFT JOIN product_variants pv ON p.dkp_id = pv.dkp_id
LEFT JOIN price_history ph ON p.dkp_id = ph.dkp_id
WHERE p.category = '$CATEGORY'
GROUP BY p.dkp_id, p.title
ORDER BY COUNT(DISTINCT pv.variant_id) DESC
LIMIT 5;
"

echo ""
echo "🔄 Next Steps:"
echo "   - Set up cron to run price-tracker 4x daily"
echo "   - Run variant-crawler weekly to discover new variants"
echo "   - Run category-crawler-v2 weekly to discover new products"
echo ""
echo "📝 Schedule Example:"
echo "   0 */6 * * * ./price-tracker          # Every 6 hours"
echo "   0 4 * * 0 ./variant-crawler          # Sunday 4 AM"
echo "   0 2 * * 0 ./category-crawler-v2      # Sunday 2 AM"
