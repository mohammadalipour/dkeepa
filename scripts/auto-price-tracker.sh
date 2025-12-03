#!/bin/bash

# Auto-run price tracker after variant crawler completes
# Usage: ./auto-price-tracker.sh

echo "🤖 Waiting for variant crawler to complete..."
echo ""

VARIANT_PID=$(ps aux | grep variant-crawler | grep -v grep | awk '{print $2}')

if [ -z "$VARIANT_PID" ]; then
    echo "✅ Variant crawler is not running"
    echo "Starting price tracker immediately..."
else
    echo "⏳ Variant crawler is running (PID: $VARIANT_PID)"
    echo "Monitoring progress..."
    echo ""
    
    # Wait for variant crawler to complete
    while ps -p $VARIANT_PID > /dev/null 2>&1; do
        # Show progress if available
        PROGRESS=$(tail -5 /tmp/variant-crawler-rerun.log 2>/dev/null | grep -o '\[[0-9]*/[0-9]*\]' | tail -1)
        if [ ! -z "$PROGRESS" ]; then
            echo -ne "\r   Progress: $PROGRESS   "
        fi
        sleep 2
    done
    
    echo ""
    echo ""
    echo "✅ Variant crawler completed!"
    sleep 2
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 Starting Price Tracker"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Build price tracker if needed
if [ ! -f ./price-tracker ]; then
    echo "📦 Building price tracker..."
    go build -o price-tracker ./cmd/price-tracker
fi

# Check how many variants we have
echo "📊 Checking database status..."
DB_STATS=$(DB_HOST=localhost ./check-db 2>&1 | grep "Trackable Variants:" -A 1)
echo "$DB_STATS"
echo ""

# Run price tracker
echo "💰 Running price tracker..."
echo "   Log: /tmp/price-tracker-auto.log"
echo ""

DB_HOST=localhost ./price-tracker > /tmp/price-tracker-auto.log 2>&1

# Check results
if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Price tracker completed successfully!"
    echo ""
    echo "📊 Final Results:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    DB_HOST=localhost ./check-db | grep -A 8 "💰 Price History:"
    echo ""
    echo "📝 View detailed log:"
    echo "   tail -f /tmp/price-tracker-auto.log"
else
    echo ""
    echo "❌ Price tracker failed!"
    echo "Check log: tail /tmp/price-tracker-auto.log"
    exit 1
fi
