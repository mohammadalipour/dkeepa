#!/bin/bash

# Keepa - Trigger Product Scrape
# Usage: ./scripts/scrape-product.sh <dkp_id>

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <dkp_id>"
    echo "Example: $0 20758981"
    exit 1
fi

DKP_ID=$1
VARIANT_ID=$2

echo "🔍 Triggering scrape for product: $DKP_ID (Variant: ${VARIANT_ID:-default})"

# Construct payload
if [ -n "$VARIANT_ID" ]; then
    PAYLOAD="{\"dkp_id\": \"$DKP_ID\", \"variant_id\": \"$VARIANT_ID\"}"
else
    PAYLOAD="{\"dkp_id\": \"$DKP_ID\"}"
fi

# Publish message to RabbitMQ
docker exec keepa-rabbitmq rabbitmqadmin publish \
    exchange=amq.default \
    routing_key=scrape_tasks \
    payload="$PAYLOAD"

echo "✅ Scrape task queued successfully!"
echo ""
echo "📊 Monitor the worker logs:"
echo "   docker logs keepa-worker -f"
echo ""
echo "🔍 Check the database after a few seconds:"
echo "   docker exec keepa-timescaledb psql -U postgres -d keepa -c \"SELECT * FROM price_history WHERE dkp_id = '$DKP_ID' ORDER BY time DESC LIMIT 5;\""
