#!/bin/bash

# Live monitoring dashboard for crawlers
# Usage: ./monitor-live.sh

clear

echo "🔍 Keepa Crawler Live Monitor"
echo "Press Ctrl+C to exit"
echo ""

while true; do
    # Move cursor to top
    tput cup 0 0
    
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║          📊 KEEPA CRAWLER LIVE DASHBOARD                    ║"
    echo "║          $(date '+%Y-%m-%d %H:%M:%S')                              ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
    
    # Check database connection
    if ! docker exec keepa-timescaledb psql -U postgres -d keepa -c "SELECT 1" > /dev/null 2>&1; then
        echo "❌ Database not available"
        sleep 5
        continue
    fi
    
    # Get statistics from database
    DB_STATS=$(docker exec keepa-timescaledb psql -U postgres -d keepa -t -A -F'|' -c "
        SELECT 
            COUNT(*) as total_products,
            COUNT(*) FILTER (WHERE is_active = true) as active_products,
            COUNT(DISTINCT category) as categories,
            MAX(last_crawled) as last_product_crawl
        FROM products;
        
        SELECT COUNT(*) as total_variants
        FROM product_variants;
        
        SELECT COUNT(*) as total_prices
        FROM price_history;
        
        SELECT 
            category_slug,
            product_count,
            last_crawled
        FROM categories
        ORDER BY product_count DESC
        LIMIT 5;
    ")
    
    # Parse database stats
    PRODUCT_LINE=$(echo "$DB_STATS" | sed -n '1p')
    TOTAL_PRODUCTS=$(echo "$PRODUCT_LINE" | cut -d'|' -f1)
    ACTIVE_PRODUCTS=$(echo "$PRODUCT_LINE" | cut -d'|' -f2)
    CATEGORIES=$(echo "$PRODUCT_LINE" | cut -d'|' -f3)
    LAST_PRODUCT_CRAWL=$(echo "$PRODUCT_LINE" | cut -d'|' -f4)
    
    TOTAL_VARIANTS=$(echo "$DB_STATS" | sed -n '2p')
    TOTAL_PRICES=$(echo "$DB_STATS" | sed -n '3p')
    
    # Calculate inactive products
    INACTIVE_PRODUCTS=$((TOTAL_PRODUCTS - ACTIVE_PRODUCTS))
    
    # Calculate percentages
    if [ $TOTAL_PRODUCTS -gt 0 ]; then
        ACTIVE_PERCENT=$(awk "BEGIN {printf \"%.1f\", ($ACTIVE_PRODUCTS/$TOTAL_PRODUCTS)*100}")
        INACTIVE_PERCENT=$(awk "BEGIN {printf \"%.1f\", ($INACTIVE_PRODUCTS/$TOTAL_PRODUCTS)*100}")
    else
        ACTIVE_PERCENT="0.0"
        INACTIVE_PERCENT="0.0"
    fi
    
    # Products
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│ 📦 PRODUCTS                                                 │"
    echo "├─────────────────────────────────────────────────────────────┤"
    printf "│ Total:      %-48s│\n" "$TOTAL_PRODUCTS products across $CATEGORIES categories"
    printf "│ Active:     %-48s│\n" "🟢 $ACTIVE_PRODUCTS ($ACTIVE_PERCENT%)"
    printf "│ Inactive:   %-48s│\n" "⚫ $INACTIVE_PRODUCTS ($INACTIVE_PERCENT%)"
    printf "│ Last Crawl: %-48s│\n" "${LAST_PRODUCT_CRAWL:-Never}"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo ""
    
    # Variants
    if [ $ACTIVE_PRODUCTS -gt 0 ]; then
        VARIANT_PERCENT=$(awk "BEGIN {printf \"%.1f\", ($TOTAL_VARIANTS/$ACTIVE_PRODUCTS)*100}")
    else
        VARIANT_PERCENT="0.0"
    fi
    
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│ 🎨 VARIANTS                                                 │"
    echo "├─────────────────────────────────────────────────────────────┤"
    printf "│ Total:      %-48s│\n" "$TOTAL_VARIANTS variants"
    printf "│ Coverage:   %-48s│\n" "$VARIANT_PERCENT% of active products"
    if [ $ACTIVE_PRODUCTS -gt 0 ]; then
        AVG_VARIANTS=$(awk "BEGIN {printf \"%.1f\", $TOTAL_VARIANTS/$ACTIVE_PRODUCTS}")
        printf "│ Average:    %-48s│\n" "$AVG_VARIANTS variants per product"
    fi
    echo "└─────────────────────────────────────────────────────────────┘"
    echo ""
    
    # Price History
    if [ $TOTAL_VARIANTS -gt 0 ]; then
        PRICE_PERCENT=$(awk "BEGIN {printf \"%.1f\", ($TOTAL_PRICES/$TOTAL_VARIANTS)*100}")
    else
        PRICE_PERCENT="0.0"
    fi
    
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│ 💰 PRICE HISTORY                                            │"
    echo "├─────────────────────────────────────────────────────────────┤"
    printf "│ Total:      %-48s│\n" "$TOTAL_PRICES price logs"
    printf "│ Coverage:   %-48s│\n" "$PRICE_PERCENT% of variants tracked"
    if [ $TOTAL_VARIANTS -gt 0 ]; then
        AVG_PRICES=$(awk "BEGIN {printf \"%.1f\", $TOTAL_PRICES/$TOTAL_VARIANTS}")
        printf "│ Average:    %-48s│\n" "$AVG_PRICES price points per variant"
    fi
    echo "└─────────────────────────────────────────────────────────────┘"
    echo ""
    
    # Active Crawlers
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│ 🚀 ACTIVE CRAWLERS                                          │"
    echo "├─────────────────────────────────────────────────────────────┤"
    
    CATEGORY_RUNNING=$(ps aux | grep "category-crawler" | grep -v grep | wc -l | tr -d ' ')
    VARIANT_RUNNING=$(ps aux | grep "variant-crawler" | grep -v grep | wc -l | tr -d ' ')
    PRICE_RUNNING=$(ps aux | grep "price-tracker" | grep -v grep | wc -l | tr -d ' ')
    
    if [ $CATEGORY_RUNNING -gt 0 ]; then
        echo "│ 📂 Category Crawler:  🟢 RUNNING                            │"
        if [ -f /tmp/category-crawler-v2.log ]; then
            LAST_LOG=$(tail -1 /tmp/category-crawler-v2.log | cut -c1-55)
            printf "│    %-57s│\n" "$LAST_LOG"
        fi
    else
        echo "│ 📂 Category Crawler:  ⚫ STOPPED                            │"
    fi
    
    if [ $VARIANT_RUNNING -gt 0 ]; then
        echo "│ 🎨 Variant Crawler:   🟢 RUNNING                            │"
        if [ -f /tmp/variant-crawler.log ]; then
            LAST_LOG=$(tail -1 /tmp/variant-crawler.log | grep -oP '\[\d+/\d+\]' || echo "Starting...")
            printf "│    Progress: %-47s│\n" "$LAST_LOG"
        fi
    else
        echo "│ 🎨 Variant Crawler:   ⚫ STOPPED                            │"
    fi
    
    if [ $PRICE_RUNNING -gt 0 ]; then
        echo "│ 💰 Price Tracker:     🟢 RUNNING                            │"
        if [ -f /tmp/price-tracker.log ]; then
            LAST_LOG=$(tail -1 /tmp/price-tracker.log | grep -oP '\[\d+/\d+\]' || echo "Starting...")
            printf "│    Progress: %-47s│\n" "$LAST_LOG"
        fi
    else
        echo "│ 💰 Price Tracker:     ⚫ STOPPED                            │"
    fi
    echo "└─────────────────────────────────────────────────────────────┘"
    echo ""
    
    # Categories breakdown
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│ 📋 TOP CATEGORIES                                           │"
    echo "├─────────────────────────────────────────────────────────────┤"
    
    # Get category breakdown
    docker exec keepa-timescaledb psql -U postgres -d keepa -t -A -F'|' -c "
        SELECT 
            category,
            COUNT(*) as total,
            COUNT(*) FILTER (WHERE is_active = true) as active
        FROM products
        GROUP BY category
        ORDER BY total DESC
        LIMIT 5;
    " | while IFS='|' read -r cat total active; do
        if [ ! -z "$cat" ]; then
            printf "│ %-15s Total: %-6s Active: %-6s           │\n" "$cat" "$total" "$active"
        fi
    done
    
    echo "└─────────────────────────────────────────────────────────────┘"
    echo ""
    
    # Next steps
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│ 📝 NEXT STEPS                                               │"
    echo "├─────────────────────────────────────────────────────────────┤"
    
    if [ $CATEGORY_RUNNING -eq 0 ] && [ $ACTIVE_PRODUCTS -lt 100 ]; then
        echo "│ ⚠️  Run category crawler to discover products               │"
        echo "│    ./category-crawler-v2 --category=mobile-phone --max=0   │"
    fi
    
    if [ $VARIANT_RUNNING -eq 0 ] && [ $ACTIVE_PRODUCTS -gt 0 ] && [ $TOTAL_VARIANTS -eq 0 ]; then
        echo "│ ⚠️  Run variant crawler to discover variants                │"
        echo "│    ./variant-crawler                                       │"
    fi
    
    if [ $PRICE_RUNNING -eq 0 ] && [ $TOTAL_VARIANTS -gt 0 ]; then
        echo "│ ⚠️  Run price tracker to collect prices                     │"
        echo "│    ./price-tracker                                         │"
    fi
    
    if [ $CATEGORY_RUNNING -eq 0 ] && [ $VARIANT_RUNNING -eq 0 ] && [ $PRICE_RUNNING -eq 0 ] && [ $TOTAL_PRICES -gt 0 ]; then
        echo "│ ✅ All crawlers stopped. System ready!                      │"
        echo "│    Prices are being collected. Check again in 6 hours.    │"
    fi
    
    echo "└─────────────────────────────────────────────────────────────┘"
    echo ""
    echo "Refreshing in 5 seconds... (Ctrl+C to exit)"
    
    sleep 5
done
