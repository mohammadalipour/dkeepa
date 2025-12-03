#!/bin/bash

# Simple progress checker that uses local psql or checks logs
# Usage: ./check-progress.sh

echo "🔍 Keepa Crawler Progress Report"
echo "================================="
echo ""

# Check variant crawler log
if [ -f /tmp/variant-crawler.log ]; then
    echo "📊 Variant Crawler Status:"
    echo "-------------------------"
    
    # Get last few lines showing progress
    LAST_LINES=$(tail -5 /tmp/variant-crawler.log)
    echo "$LAST_LINES"
    echo ""
    
    # Try to extract progress numbers (macOS compatible)
    PROGRESS=$(echo "$LAST_LINES" | grep -o '\[[0-9]*/[0-9]*\]' | tail -1)
    if [ ! -z "$PROGRESS" ]; then
        CURRENT=$(echo "$PROGRESS" | sed 's/\[//' | cut -d'/' -f1)
        TOTAL=$(echo "$PROGRESS" | sed 's/\]//' | cut -d'/' -f2)
        
        if [ ! -z "$CURRENT" ] && [ ! -z "$TOTAL" ]; then
            PERCENT=$(awk "BEGIN {printf \"%.1f\", ($CURRENT/$TOTAL)*100}")
            echo "   Progress: $CURRENT/$TOTAL products ($PERCENT%)"
            
            REMAINING=$((TOTAL - CURRENT))
            echo "   Remaining: $REMAINING products"
            
            # Estimate time (assume 2 seconds per product)
            TIME_LEFT=$((REMAINING * 2))
            MINUTES=$((TIME_LEFT / 60))
            SECONDS=$((TIME_LEFT % 60))
            echo "   Estimated time left: ${MINUTES}m ${SECONDS}s"
        fi
    fi
    echo ""
else
    echo "⚠️  Variant crawler log not found at /tmp/variant-crawler.log"
    echo ""
fi

# Check if process is running
VARIANT_PID=$(ps aux | grep variant-crawler | grep -v grep | awk '{print $2}')
if [ ! -z "$VARIANT_PID" ]; then
    echo "✅ Variant crawler is RUNNING (PID: $VARIANT_PID)"
else
    echo "⚫ Variant crawler is STOPPED"
fi
echo ""

# Check category crawler log
if [ -f /tmp/category-crawler-v2.log ]; then
    echo "📊 Category Crawler Status:"
    echo "--------------------------"
    tail -3 /tmp/category-crawler-v2.log
    echo ""
fi

# Check price tracker log
if [ -f /tmp/price-tracker.log ]; then
    echo "📊 Price Tracker Status:"
    echo "------------------------"
    tail -3 /tmp/price-tracker.log
    echo ""
fi

echo ""
echo "📊 Database Status:"
echo "-------------------"

# Get database stats using the check-db tool
if [ -f ./check-db ]; then
    DB_HOST=localhost ./check-db 2>/dev/null | grep -A 20 "Database Diagnostics" | tail -15
else
    echo "   ⚠️  Run 'go build -o check-db ./cmd/check-db' to see detailed stats"
fi

echo ""
echo "💡 Tips:"
echo "  • Monitor in real-time: tail -f /tmp/variant-crawler.log"
echo "  • Check database: psql -h localhost -U postgres -d keepa"
echo "  • View process: ps aux | grep crawler"
echo ""
