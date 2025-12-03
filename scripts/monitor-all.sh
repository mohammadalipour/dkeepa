#!/bin/bash

# Comprehensive monitoring script for all Keepa components
# Usage: ./monitor-all.sh [--loop]

LOOP_MODE=false
if [ "$1" = "--loop" ]; then
    LOOP_MODE=true
fi

function show_status() {
    clear
    
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║          📊 KEEPA PRICE TRACKER - FULL STATUS               ║"
    echo "║          $(date '+%Y-%m-%d %H:%M:%S')                              ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
    
    # Check if database tools are available
    if [ ! -f ./check-db ]; then
        echo "⚠️  Building database checker..."
        go build -o check-db ./cmd/check-db 2>&1 | grep -v "^#" || true
        echo ""
    fi
    
    # Get database statistics
    if [ -f ./check-db ]; then
        echo "┌─────────────────────────────────────────────────────────────┐"
        echo "│ 📦 DATABASE STATISTICS                                      │"
        echo "└─────────────────────────────────────────────────────────────┘"
        
        DB_OUTPUT=$(DB_HOST=localhost ./check-db 2>&1)
        
        # Products
        TOTAL_PRODUCTS=$(echo "$DB_OUTPUT" | grep "Total:" | head -1 | awk '{print $2}')
        ACTIVE_PRODUCTS=$(echo "$DB_OUTPUT" | grep "Active:" | head -1 | awk '{print $2}')
        TRACKED_PRODUCTS=$(echo "$DB_OUTPUT" | grep "Tracked:" | head -1 | awk '{print $2}')
        
        echo "  📦 Products:"
        echo "     Total:   $TOTAL_PRODUCTS"
        echo "     Active:  $ACTIVE_PRODUCTS (marketable)"
        echo "     Tracked: $TRACKED_PRODUCTS"
        echo ""
        
        # Variants
        TOTAL_VARIANTS=$(echo "$DB_OUTPUT" | grep "Total:" | tail -2 | head -1 | awk '{print $2}')
        ACTIVE_VARIANTS=$(echo "$DB_OUTPUT" | grep "Active:" | tail -2 | head -1 | awk '{print $2}')
        
        echo "  🎨 Variants:"
        echo "     Total:  $TOTAL_VARIANTS"
        echo "     Active: $ACTIVE_VARIANTS (trackable)"
        
        if [ ! -z "$ACTIVE_PRODUCTS" ] && [ "$ACTIVE_PRODUCTS" != "0" ]; then
            AVG_VARIANTS=$(awk "BEGIN {printf \"%.1f\", $TOTAL_VARIANTS/$ACTIVE_PRODUCTS}")
            echo "     Average: $AVG_VARIANTS per product"
        fi
        echo ""
        
        # Price History
        PRICE_ENTRIES=$(echo "$DB_OUTPUT" | grep "Total Entries:" | awk '{print $3}')
        UNIQUE_PRODUCTS=$(echo "$DB_OUTPUT" | grep "Unique Variants:" | awk '{print $3}')
        LATEST_PRICE=$(echo "$DB_OUTPUT" | grep "Latest Price Time:" | cut -d: -f2- | xargs)
        
        echo "  💰 Price History:"
        echo "     Total Entries: $PRICE_ENTRIES"
        echo "     Unique Products: $UNIQUE_PRODUCTS"
        echo "     Latest Price: $LATEST_PRICE"
        
        # Calculate coverage
        if [ ! -z "$ACTIVE_VARIANTS" ] && [ "$ACTIVE_VARIANTS" != "0" ] && [ ! -z "$UNIQUE_PRODUCTS" ] && [ "$UNIQUE_PRODUCTS" != "0" ]; then
            COVERAGE=$(awk "BEGIN {printf \"%.1f\", ($UNIQUE_PRODUCTS/$ACTIVE_VARIANTS)*100}")
            echo "     Coverage: $COVERAGE% of active variants"
        fi
        
        # Calculate freshness (prices from last 24 hours)
        RECENT_PRICES=$(DB_HOST=localhost ./check-db 2>&1 | grep -o "Total Entries:.*" | awk '{print $3}')
        if [ ! -z "$RECENT_PRICES" ]; then
            echo "     Status: 🟢 Data available"
        else
            echo "     Status: ⚪ No recent data"
        fi
        echo ""
        
    else
        echo "❌ Unable to check database (check-db tool not available)"
        echo ""
    fi
    
    # Crawler status
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│ 🚀 CRAWLER STATUS                                           │"
    echo "└─────────────────────────────────────────────────────────────┘"
    
    # Category Crawler
    CATEGORY_PID=$(ps aux | grep "category-crawler" | grep -v grep | awk '{print $2}')
    if [ ! -z "$CATEGORY_PID" ]; then
        echo "  📂 Category Crawler: 🟢 RUNNING (PID: $CATEGORY_PID)"
        if [ -f /tmp/category-crawler-v2.log ]; then
            LAST_CAT=$(tail -1 /tmp/category-crawler-v2.log | cut -c1-50)
            echo "     $LAST_CAT"
        fi
    else
        echo "  📂 Category Crawler: ⚫ STOPPED"
    fi
    echo ""
    
    # Variant Crawler
    VARIANT_PID=$(ps aux | grep "variant-crawler" | grep -v grep | awk '{print $2}')
    if [ ! -z "$VARIANT_PID" ]; then
        echo "  🎨 Variant Crawler: 🟢 RUNNING (PID: $VARIANT_PID)"
        
        # Try both log files
        if [ -f /tmp/variant-crawler-rerun.log ]; then
            PROGRESS=$(tail -5 /tmp/variant-crawler-rerun.log | grep -o '\[[0-9]*/[0-9]*\]' | tail -1)
            if [ ! -z "$PROGRESS" ]; then
                CURRENT=$(echo "$PROGRESS" | sed 's/\[//' | cut -d'/' -f1)
                TOTAL=$(echo "$PROGRESS" | sed 's/\]//' | cut -d'/' -f2)
                PERCENT=$(awk "BEGIN {printf \"%.1f\", ($CURRENT/$TOTAL)*100}")
                REMAINING=$((TOTAL - CURRENT))
                TIME_LEFT=$((REMAINING * 2))
                MINUTES=$((TIME_LEFT / 60))
                SECONDS=$((TIME_LEFT % 60))
                
                echo "     Progress: $CURRENT/$TOTAL ($PERCENT%)"
                echo "     Remaining: $REMAINING products"
                echo "     ETA: ${MINUTES}m ${SECONDS}s"
            fi
        elif [ -f /tmp/variant-crawler.log ]; then
            PROGRESS=$(tail -5 /tmp/variant-crawler.log | grep -o '\[[0-9]*/[0-9]*\]' | tail -1)
            if [ ! -z "$PROGRESS" ]; then
                echo "     Progress: $PROGRESS"
            fi
        fi
    else
        echo "  🎨 Variant Crawler: ⚫ STOPPED"
    fi
    echo ""
    
    # Price Tracker
    PRICE_PID=$(ps aux | grep "price-tracker" | grep -v grep | awk '{print $2}')
    if [ ! -z "$PRICE_PID" ]; then
        echo "  💰 Price Tracker: 🟢 RUNNING (PID: $PRICE_PID)"
        if [ -f /tmp/price-tracker.log ]; then
            LAST_PRICE=$(tail -1 /tmp/price-tracker.log | cut -c1-50)
            echo "     $LAST_PRICE"
        fi
    else
        echo "  💰 Price Tracker: ⚫ STOPPED"
    fi
    echo ""
    
    # Recommendations
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│ 💡 RECOMMENDATIONS                                          │"
    echo "└─────────────────────────────────────────────────────────────┘"
    
    if [ -z "$CATEGORY_PID" ] && [ -z "$ACTIVE_PRODUCTS" -o "$ACTIVE_PRODUCTS" = "0" ]; then
        echo "  ⚠️  No products found. Run category crawler:"
        echo "     DB_HOST=localhost ./category-crawler-v2 --category=mobile-phone --max=0"
        echo ""
    fi
    
    if [ -z "$VARIANT_PID" ] && [ ! -z "$ACTIVE_PRODUCTS" ] && [ "$ACTIVE_PRODUCTS" != "0" ] && [ "$ACTIVE_VARIANTS" = "0" ]; then
        echo "  ⚠️  No variants found. Run variant crawler:"
        echo "     DB_HOST=localhost ./variant-crawler"
        echo ""
    fi
    
    if [ -z "$PRICE_PID" ] && [ ! -z "$ACTIVE_VARIANTS" ] && [ "$ACTIVE_VARIANTS" != "0" ]; then
        if [ "$UNIQUE_PRODUCTS" = "0" -o -z "$UNIQUE_PRODUCTS" ]; then
            echo "  ⚠️  No prices tracked yet. Run price tracker:"
            echo "     DB_HOST=localhost ./price-tracker"
            echo ""
        elif [ ! -z "$COVERAGE" ]; then
            if (( $(echo "$COVERAGE < 80" | bc -l) )); then
                echo "  ⚠️  Low price coverage ($COVERAGE%). Consider running:"
                echo "     DB_HOST=localhost ./price-tracker"
                echo ""
            fi
        fi
    fi
    
    if [ -z "$CATEGORY_PID" ] && [ -z "$VARIANT_PID" ] && [ -z "$PRICE_PID" ]; then
        if [ ! -z "$PRICE_ENTRIES" ] && [ "$PRICE_ENTRIES" != "0" ]; then
            echo "  ✅ All crawlers stopped. System is ready!"
            echo "     Set up cron jobs for automated tracking:"
            echo "     - Prices: Every 6 hours"
            echo "     - Variants: Weekly"
            echo "     - Products: Weekly"
            echo ""
        fi
    fi
    
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│ 📝 QUICK COMMANDS                                           │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo "  View logs:"
    echo "    tail -f /tmp/variant-crawler-rerun.log"
    echo "    tail -f /tmp/price-tracker.log"
    echo ""
    echo "  Manual runs:"
    echo "    DB_HOST=localhost ./variant-crawler"
    echo "    DB_HOST=localhost ./price-tracker"
    echo ""
}

# Main loop
if [ "$LOOP_MODE" = true ]; then
    echo "🔄 Starting continuous monitoring (Ctrl+C to exit)..."
    while true; do
        show_status
        echo "Refreshing in 10 seconds..."
        sleep 10
    done
else
    show_status
    echo "💡 Tip: Run with --loop for continuous monitoring"
    echo ""
fi
