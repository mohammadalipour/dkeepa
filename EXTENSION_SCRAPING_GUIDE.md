# Extension-Based Product Scraping Guide

## Overview

The Keepa extension now scrapes product data directly from Digikala pages and sends it to the backend. This solves the `_rch` parameter challenge and provides real-time data collection as users browse.

## How It Works

### 1. **Extension Content Script** (`extension/src/content/`)

When a user visits a Digikala product page:

#### `digikalaScraper.ts` - Core Scraping Logic
- **Extracts `_rch` token** from:
  - URL parameters (`?_rch=...`)
  - Next.js `__NEXT_DATA__` script tag
  - Inline scripts
  - Intercepted API requests (Fetch API hook)
  
- **Fetches product data** from Digikala API:
  - Uses extracted `_rch` token
  - Calls `https://api.digikala.com/v2/product/{dkp_id}/?_rch={token}&variant_id={vid}`
  - Parses JSON response for price, title, seller, etc.
  
- **Fallback to DOM parsing** if API fails:
  - Extracts from JSON-LD structured data
  - Parses price/title from page elements

#### `index.tsx` - Extension Entry Point
- Detects product ID and variant ID from URL
- Waits for page to load (1-2 seconds)
- Calls `scrapeProductData()`
- Sends scraped data to background script
- Injects price chart UI

### 2. **Background Script** (`extension/src/background/index.ts`)

- Receives `SEND_PRODUCT_DATA` message from content script
- Forwards data to backend API endpoint: `POST /api/v1/products/ingest`
- Handles API errors and provides feedback

### 3. **Backend API** (`internal/adapters/http/`)

#### New Endpoint: `POST /api/v1/products/ingest`

**Handler:** `handlers/price_handler.go::IngestProductData`

**Request Body:**
```json
{
  "dkp_id": "20758981",
  "variant_id": "73675351",
  "title": "لپ تاپ 15.6 اینچی لنوو...",
  "price": 306990000,
  "rrp_price": 325000000,
  "seller_name": "دیجی‌کالا",
  "is_active": true,
  "rch_token": "89e796febf77"
}
```

**Flow:**
1. Validates request data
2. Creates `domain.Product` and `domain.PriceLog` objects
3. Calls `PriceService.SaveProductPrice()`
4. Service upserts product and inserts price log
5. Returns success response

## Data Flow

```
User visits Digikala product page
         ↓
Extension content script loads
         ↓
Extract _rch token + product data
         ↓
Scrape from Digikala API (with _rch)
         ↓
Send to background script
         ↓
POST to backend /api/v1/products/ingest
         ↓
Backend saves to TimescaleDB
         ↓
Extension displays price chart
```

## Advantages

✅ **No anti-bot issues** - Runs in real browser with user context  
✅ **Always has valid `_rch`** - Extracted from live page  
✅ **Real-time data** - Captured as users browse  
✅ **Crowdsourced** - Users help collect data naturally  
✅ **Fallback support** - DOM parsing if API fails  

## Testing

### 1. Build the Extension
```bash
cd extension
npm install
npm run build
```

### 2. Load in Chrome
1. Open `chrome://extensions`
2. Enable "Developer mode"
3. Click "Load unpacked"
4. Select `extension/dist` folder

### 3. Start Backend
```bash
# Start database
docker-compose up -d timescaledb

# Start API server
go run cmd/api/main.go
```

### 4. Visit a Digikala Product Page
Example: `https://www.digikala.com/product/dkp-20758981/?variant_id=73675351`

### 5. Check Logs
- **Browser Console**: See scraping debug info
- **Backend Logs**: See ingestion requests
- **Database**: Query `products` and `price_history` tables

```sql
SELECT * FROM products WHERE dkp_id = '20758981';
SELECT * FROM price_history WHERE dkp_id = '20758981' ORDER BY time DESC LIMIT 10;
```

## Troubleshooting

### "_rch token not found"
- Check browser console for extraction attempts
- Verify you're on a product page (URL contains `/product/dkp-`)
- Wait 2-3 seconds for page to fully load
- Check if Digikala changed their page structure

### "Failed to send to backend"
- Ensure API server is running on `localhost:8080`
- Check CORS settings (should allow extension origin)
- Check network tab for error details
- Verify database is accessible

### "API returned 404"
- `_rch` token may have expired (refresh page)
- Product ID might be invalid
- Digikala API structure may have changed

## Future Enhancements

1. **Automated Monitoring** - Worker scrapes hot products periodically
2. **Price Alerts** - Notify users when prices drop
3. **Multi-seller Tracking** - Track all sellers, not just buy box
4. **Historical Comparison** - Show price trends over time
5. **Export Data** - Allow users to export price history

## Files Modified

### Extension
- ✅ `extension/src/content/digikalaScraper.ts` (NEW)
- ✅ `extension/src/content/index.tsx`
- ✅ `extension/src/background/index.ts`

### Backend
- ✅ `internal/adapters/http/handlers/price_handler.go`
- ✅ `internal/adapters/http/router.go`
- ✅ `internal/core/ports/service.go`
- ✅ `internal/core/services/price_service.go`

### Infrastructure
- ✅ TLS Client cookie handling fixed (no redirect loop)
- ✅ Warmup request to establish cookies

## Next Steps

After the extension is working well with user-browsing data, we can add:
1. **Go scraper fallback** for automated monitoring
2. **Scheduler** to periodically check hot products
3. **RabbitMQ integration** for distributed scraping tasks
