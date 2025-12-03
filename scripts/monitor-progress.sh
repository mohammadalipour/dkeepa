#!/bin/bash

# Keepa Crawler Progress Monitor
# Shows real-time progress of product, variant, and price tracking

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Database connection
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-keepa}"

# Function to execute SQL and get result
query_db() {
    docker exec keepa-timescaledb psql -U "$DB_USER" -d "$DB_NAME" -t -A -c "$1"
}

# Clear screen and show header
clear
echo -e "${BOLD}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║          KEEPA CRAWLER PROGRESS MONITOR                    ║${NC}"
echo -e "${BOLD}╔════════════════════════════════════════════════════════════╗${NC}"
echo ""

# 1. CATEGORY CRAWLER PROGRESS
echo -e "${BLUE}${BOLD}📊 CATEGORY CRAWLER (Product Discovery)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Get category stats
category_stats=$(query_db "
SELECT 
    category,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE is_active) as active,
    COUNT(*) FILTER (WHERE is_tracked) as tracked,
    MAX(last_crawled) as last_crawled
FROM products 
WHERE category IS NOT NULL
GROUP BY category
ORDER BY total DESC;
")

if [ -z "$category_stats" ]; then
    echo -e "${YELLOW}⚠️  No products found. Run category crawler first.${NC}"
else
    echo -e "${BOLD}Category               Total    Active   Tracked   Last Crawled${NC}"
    echo "$category_stats" | while IFS='|' read -r cat total active tracked last_crawl; do
        printf "%-20s %6s %8s %9s   %s\n" "$cat" "$total" "$active" "$tracked" "${last_crawl:0:19}"
    done
    
    # Calculate total
    total_products=$(query_db "SELECT COUNT(*) FROM products WHERE category IS NOT NULL;")
    echo ""
    echo -e "${GREEN}✅ Total Products Discovered: ${BOLD}$total_products${NC}"
fi

echo ""

# 2. VARIANT CRAWLER PROGRESS
echo -e "${BLUE}${BOLD}🔍 VARIANT CRAWLER (Variant Discovery)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Calculate variant progress
variant_progress=$(query_db "
SELECT 
    COUNT(DISTINCT p.dkp_id) as total_products,
    COUNT(DISTINCT pv.dkp_id) as products_with_variants,
    COUNT(pv.variant_id) as total_variants,
    ROUND(100.0 * COUNT(DISTINCT pv.dkp_id) / NULLIF(COUNT(DISTINCT p.dkp_id), 0), 2) as percentage
FROM products p
LEFT JOIN product_variants pv ON p.dkp_id = pv.dkp_id
WHERE p.is_tracked = true AND p.category IS NOT NULL;
")

IFS='|' read -r total_prods prods_w_variants total_vars percentage <<< "$variant_progress"

if [ "$total_prods" = "0" ]; then
    echo -e "${YELLOW}⚠️  No tracked products found.${NC}"
else
    echo -e "Total Tracked Products:      ${BOLD}$total_prods${NC}"
    echo -e "Products with Variants:      ${BOLD}$prods_w_variants${NC}"
    echo -e "Total Variants Discovered:   ${BOLD}$total_vars${NC}"
    echo ""
    
    # Progress bar
    percentage_int=${percentage%.*}
    if [ -z "$percentage_int" ]; then
        percentage_int=0
    fi
    
    filled=$((percentage_int / 2))
    empty=$((50 - filled))
    
    printf "Progress: ["
    printf "%${filled}s" | tr ' ' '█'
    printf "%${empty}s" | tr ' ' '░'
    printf "] ${BOLD}${percentage}%%${NC}\n"
    
    if [ "$percentage_int" -eq 100 ]; then
        echo -e "${GREEN}✅ All products have variants discovered!${NC}"
    else
        remaining=$((total_prods - prods_w_variants))
        echo -e "${YELLOW}⏳ Remaining: $remaining products${NC}"
    fi
fi

echo ""

# 3. PRICE TRACKER PROGRESS
echo -e "${BLUE}${BOLD}💰 PRICE TRACKER (Price Collection)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Calculate price tracking progress
price_progress=$(query_db "
SELECT 
    COUNT(DISTINCT pv.variant_id) as total_variants,
    COUNT(DISTINCT ph.variant_id) as variants_with_prices,
    COUNT(ph.time) as total_price_records,
    ROUND(100.0 * COUNT(DISTINCT ph.variant_id) / NULLIF(COUNT(DISTINCT pv.variant_id), 0), 2) as percentage,
    MAX(ph.time) as last_price_update
FROM product_variants pv
LEFT JOIN price_history ph ON pv.variant_id::text = ph.variant_id
WHERE pv.is_active = true;
")

IFS='|' read -r total_variants vars_w_prices total_prices price_pct last_update <<< "$price_progress"

if [ "$total_variants" = "0" ]; then
    echo -e "${YELLOW}⚠️  No variants found. Run variant crawler first.${NC}"
else
    echo -e "Total Active Variants:         ${BOLD}$total_variants${NC}"
    echo -e "Variants with Prices:          ${BOLD}$vars_w_prices${NC}"
    echo -e "Total Price Records:           ${BOLD}$total_prices${NC}"
    echo -e "Last Price Update:             ${last_update:0:19}"
    echo ""
    
    # Progress bar
    price_pct_int=${price_pct%.*}
    if [ -z "$price_pct_int" ]; then
        price_pct_int=0
    fi
    
    filled=$((price_pct_int / 2))
    empty=$((50 - filled))
    
    printf "Progress: ["
    printf "%${filled}s" | tr ' ' '█'
    printf "%${empty}s" | tr ' ' '░'
    printf "] ${BOLD}${price_pct}%%${NC}\n"
    
    if [ "$price_pct_int" -eq 100 ]; then
        echo -e "${GREEN}✅ All variants have prices tracked!${NC}"
    else
        remaining=$((total_variants - vars_w_prices))
        echo -e "${YELLOW}⏳ Remaining: $remaining variants${NC}"
    fi
    
    # Average prices per variant
    if [ "$vars_w_prices" != "0" ]; then
        avg_prices=$((total_prices / vars_w_prices))
        echo -e "Average price records per variant: ${BOLD}$avg_prices${NC}"
    fi
fi

echo ""

# 4. PER-CATEGORY BREAKDOWN
echo -e "${BLUE}${BOLD}📂 PER-CATEGORY BREAKDOWN${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

category_breakdown=$(query_db "
SELECT 
    p.category,
    COUNT(DISTINCT p.dkp_id) as products,
    COUNT(DISTINCT pv.variant_id) as variants,
    COUNT(DISTINCT ph.variant_id) as priced_variants,
    CASE 
        WHEN COUNT(DISTINCT pv.variant_id) > 0 
        THEN ROUND(100.0 * COUNT(DISTINCT ph.variant_id) / COUNT(DISTINCT pv.variant_id), 1)
        ELSE 0 
    END as price_coverage
FROM products p
LEFT JOIN product_variants pv ON p.dkp_id = pv.dkp_id
LEFT JOIN price_history ph ON pv.variant_id::text = ph.variant_id
WHERE p.category IS NOT NULL
GROUP BY p.category
ORDER BY products DESC;
")

if [ -n "$category_breakdown" ]; then
    printf "${BOLD}%-20s %8s %10s %10s %10s${NC}\n" "Category" "Products" "Variants" "Priced" "Coverage"
    echo "$category_breakdown" | while IFS='|' read -r cat prods vars priced coverage; do
        printf "%-20s %8s %10s %10s %9s%%\n" "$cat" "$prods" "$vars" "$priced" "$coverage"
    done
fi

echo ""

# 5. PIPELINE STATUS
echo -e "${BLUE}${BOLD}🔄 PIPELINE STATUS${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Check what needs to be done
products_count=$(query_db "SELECT COUNT(*) FROM products WHERE category IS NOT NULL;")
products_with_variants=$(query_db "SELECT COUNT(DISTINCT dkp_id) FROM product_variants;")
variants_with_prices=$(query_db "SELECT COUNT(DISTINCT variant_id) FROM price_history;")
total_variants=$(query_db "SELECT COUNT(*) FROM product_variants WHERE is_active = true;")

# Stage 1: Products
if [ "$products_count" = "0" ]; then
    echo -e "❌ Stage 1: Product Discovery     ${RED}NOT STARTED${NC}"
    echo -e "   → Run: ${BOLD}./category-crawler-v2 --category=mobile-phone${NC}"
else
    echo -e "✅ Stage 1: Product Discovery     ${GREEN}COMPLETED${NC} ($products_count products)"
fi

# Stage 2: Variants
if [ "$products_with_variants" = "0" ]; then
    echo -e "❌ Stage 2: Variant Discovery     ${RED}NOT STARTED${NC}"
    echo -e "   → Run: ${BOLD}./variant-crawler${NC}"
elif [ "$products_with_variants" -lt "$products_count" ]; then
    pct=$((100 * products_with_variants / products_count))
    echo -e "⏳ Stage 2: Variant Discovery     ${YELLOW}IN PROGRESS${NC} ($pct%)"
    echo -e "   → Run: ${BOLD}./variant-crawler${NC} to continue"
else
    echo -e "✅ Stage 2: Variant Discovery     ${GREEN}COMPLETED${NC} ($total_variants variants)"
fi

# Stage 3: Prices
if [ "$variants_with_prices" = "0" ]; then
    echo -e "❌ Stage 3: Price Tracking        ${RED}NOT STARTED${NC}"
    echo -e "   → Run: ${BOLD}./price-tracker${NC}"
elif [ "$variants_with_prices" -lt "$total_variants" ]; then
    pct=$((100 * variants_with_prices / total_variants))
    echo -e "⏳ Stage 3: Price Tracking        ${YELLOW}IN PROGRESS${NC} ($pct%)"
    echo -e "   → Run: ${BOLD}./price-tracker${NC} to continue"
else
    echo -e "✅ Stage 3: Price Tracking        ${GREEN}COMPLETED${NC} (all variants tracked)"
fi

echo ""
echo -e "${BOLD}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
