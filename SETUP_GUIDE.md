# Keepa - Digikala Price Tracker
## Complete Step-by-Step Setup Guide

This guide will walk you through setting up and running the entire project from scratch.

---

## Prerequisites

Before starting, ensure you have:
- **Docker Desktop** installed and running
- **Node.js 20+** installed
- **Git** installed
- **Chrome Browser** (for the extension)

---

## Step 1: Start Backend Services

### 1.1 Navigate to Project Directory
```bash
cd /Users/mohammadalipour/Project/keepa
```

### 1.2 Start All Services with Docker Compose
```bash
make up
```

This will start:
- TimescaleDB (PostgreSQL with time-series)
- Redis (caching)
- RabbitMQ (message queue)
- Backend API (Go)
- Worker (scraper)

**Wait 30-60 seconds** for all services to initialize.

### 1.3 Verify Services are Running
```bash
docker ps
```

You should see 5 containers running:
- `keepa-backend`
- `keepa-worker`
- `keepa-timescaledb`
- `keepa-redis`
- `keepa-rabbitmq`

### 1.4 Test Backend API
```bash
curl http://localhost:8080/health
```

Expected response: `{"status":"ok"}`

### 1.5 Test Price History Endpoint
```bash
curl "http://localhost:8080/api/v1/products/12345/history"
```

Expected response (empty data is normal):
```json
{
  "dkp_id": "12345",
  "columns": ["time", "price", "seller_id", "is_buy_box"],
  "data": []
}
```

---

## Step 2: Build Chrome Extension

### 2.1 Navigate to Extension Directory
```bash
cd extension
```

### 2.2 Install Dependencies
```bash
npm install
```

This will install:
- React 18
- TypeScript
- Vite
- @crxjs/vite-plugin
- Recharts
- date-fns

**Wait for installation to complete** (may take 1-2 minutes).

### 2.3 Build Extension
```bash
npm run build
```

This creates the `dist` folder with the compiled extension.

---

## Step 3: Load Extension in Chrome

### 3.1 Open Chrome Extensions Page
1. Open Chrome browser
2. Navigate to: `chrome://extensions/`
3. Enable **"Developer mode"** (toggle in top-right corner)

### 3.2 Load Unpacked Extension
1. Click **"Load unpacked"** button
2. Navigate to: `/Users/mohammadalipour/Project/keepa/extension/dist`
3. Click **"Select"**

### 3.3 Verify Extension is Loaded
You should see:
- **Keepa - Digikala Price Tracker** in the extensions list
- Extension icon in Chrome toolbar
- Status: **Enabled**

---

## Step 4: Test the Extension

### 4.1 Navigate to a Digikala Product Page
Open any Digikala product URL, for example:
```
https://www.digikala.com/product/dkp-12345/
```

**Note:** Use a real product ID from Digikala for actual testing.

### 4.2 What You Should See
1. **Price Chart Widget** appears on the page (near the price section)
2. Widget shows:
   - Title: "📊 تاریخچه قیمت - Keepa"
   - Message: "هنوز داده‌ای برای این محصول ثبت نشده است" (no data yet)

### 4.3 Click Extension Icon
Click the Keepa icon in Chrome toolbar to see:
- Extension status
- Current page detection
- Version info

---

## Step 5: Populate Data (Optional)

To see actual price charts, you need to scrape some data.

### 5.1 Access RabbitMQ Management UI
Open in browser:
```
http://localhost:15672
```

Login:
- Username: `guest`
- Password: `guest`

### 5.2 Publish a Scrape Task
1. Go to **"Queues"** tab
2. Click on **"scrape_tasks"** queue
3. Scroll to **"Publish message"** section
4. In the **Payload** field, enter:
```json
{"dkp_id": "12345"}
```
5. Click **"Publish message"**

### 5.3 Monitor Worker Logs
```bash
docker logs keepa-worker -f
```

You should see:
- "Processing task: {DkpID:12345}"
- Scraping attempt (may fail if product doesn't exist)

### 5.4 Check Database
```bash
docker exec keepa-timescaledb psql -U postgres -d keepa -c "SELECT * FROM price_history LIMIT 5;"
```

---

## Step 6: Development Mode (Optional)

For extension development with hot reload:

### 6.1 Start Dev Server
```bash
cd extension
npm run dev
```

### 6.2 Load Extension from Dev Build
1. `chrome://extensions/`
2. Remove the production build
3. Load unpacked from `extension/dist` (auto-updates on changes)

---

## Troubleshooting

### Backend Not Starting
```bash
# Check logs
docker logs keepa-backend

# Restart services
make down
make up
```

### Extension Not Loading
```bash
# Rebuild extension
cd extension
rm -rf dist node_modules
npm install
npm run build
```

### Database Connection Issues
```bash
# Reset database (WARNING: deletes all data)
docker-compose down -v
make up
```

### Worker Not Scraping
```bash
# Check worker logs
docker logs keepa-worker -f

# Verify RabbitMQ connection
docker logs keepa-rabbitmq
```

---

## Useful Commands

### Backend
```bash
make up          # Start all services
make down        # Stop all services
make logs        # View all logs
docker ps        # Check running containers
```

### Extension
```bash
npm run dev      # Development mode
npm run build    # Production build
npm run preview  # Preview build
```

### Database
```bash
# Connect to database
docker exec -it keepa-timescaledb psql -U postgres -d keepa

# View tables
\dt

# View price history
SELECT * FROM price_history LIMIT 10;

# Exit
\q
```

---

## Architecture Overview

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Chrome     │────▶│   Backend    │────▶│  TimescaleDB │
│  Extension   │     │   (Gin API)  │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
                            │                      ▲
                            ▼                      │
                     ┌──────────────┐     ┌────────────┐
                     │   RabbitMQ   │────▶│   Worker   │
                     └──────────────┘     └────────────┘
```

---

## Next Steps

1. **Add Real Product IDs**: Test with actual Digikala products
2. **Monitor Scheduler**: Backend automatically queues hot products every 5 minutes
3. **Customize Extension**: Modify UI in `extension/src/content/`
4. **Scale Workers**: Add more worker instances in `docker-compose.yml`

---

## Support

For issues or questions:
1. Check logs: `make logs`
2. Verify services: `docker ps`
3. Review documentation in `README.md`

---

**You're all set! 🎉**

The Keepa - Digikala Price Tracker is now running and ready to track prices!
