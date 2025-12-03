#!/bin/bash

# Quick database diagnostics
# Usage: ./check-db-data.sh

echo "🔍 Database Diagnostics"
echo "======================="
echo ""

# Check if we have psql
if ! command -v psql &> /dev/null; then
    echo "❌ psql not found. Please install PostgreSQL client."
    echo ""
    echo "You can check the data by running Go code or connecting via Docker:"
    echo "  docker exec keepa-timescaledb psql -U postgres -d keepa -c 'SELECT COUNT(*) FROM product_variants;'"
    exit 1
fi

# Database connection
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-keepa}"

echo "📊 Products:"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT 
        COUNT(*) as total,
        COUNT(*) FILTER (WHERE is_active = true) as active,
        COUNT(*) FILTER (WHERE is_tracked = true) as tracked
    FROM products;
"

echo ""
echo "🎨 Product Variants:"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT 
        COUNT(*) as total,
        COUNT(*) FILTER (WHERE is_active = true) as active
    FROM product_variants;
"

echo ""
echo "💰 Price History:"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT 
        COUNT(*) as total,
        COUNT(DISTINCT variant_id) as unique_variants,
        MAX(timestamp) as latest_price
    FROM price_history;
"

echo ""
echo "🔍 Sample Variants (first 5):"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT 
        v.variant_id,
        v.dkp_id,
        v.variant_title,
        v.is_active,
        p.is_tracked,
        p.is_active as product_active
    FROM product_variants v
    JOIN products p ON v.dkp_id = p.dkp_id
    LIMIT 5;
"
