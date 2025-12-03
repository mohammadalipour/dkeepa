# Quick Fix: Restart Backend to Enable Multi-Variant Charts

## Current Situation
- ✅ Backend code updated to support separate variant lines
- ❌ Docker container still running old code
- ❌ Chart shows only single line

## 🚀 Quick Solution (Choose One)

### Option 1: Using Docker Desktop UI
1. Open **Docker Desktop** application
2. Find the **keepa** containers
3. Click **Restart** on the backend container
4. Wait 5 seconds
5. Refresh the Digikala page

### Option 2: Using Terminal with Docker
If you have a terminal with Docker access:

```bash
# Find the container
docker ps | grep keepa

# Restart it (replace CONTAINER_ID with actual ID)
docker restart CONTAINER_ID

# Or rebuild and restart all
docker-compose down
docker-compose build
docker-compose up -d
```

### Option 3: Check docker-compose command location
```bash
# Try these:
docker compose down  # Note: no dash (newer Docker)
docker compose build backend
docker compose up -d

# Or find it
which docker-compose
/usr/local/bin/docker-compose down
```

## 🧪 Verify It's Working

After restarting, test the API:

```bash
# Should now show "variants" array in response
curl "http://localhost:8080/api/v1/products/12017236/history" | grep -o '"variants"'
```

If you see `"variants"`, it's working! ✅

## 📊 Test Products

These products have multiple variants (perfect for testing):

1. **Product 12017236** - Has 21 variants
   - https://www.digikala.com/product/dkp-12017236/
   
2. **Product 11346346** - Samsung Galaxy A34
   - https://www.digikala.com/product/dkp-11346346/

After backend restarts, visiting these pages should show **multiple colored lines** in the chart!

## 🔧 If Docker Still Won't Work

Alternative: Run backend directly (outside Docker):

```bash
# Kill the Docker backend
# (Find PID from: lsof -i :8080)

# Run the new backend directly
DB_HOST=localhost \
DB_USER=postgres \
DB_PASSWORD=password \
DB_NAME=keepa \
./keepa-api
```

This will run the updated code on port 8080 immediately!

## 📝 Expected Result

After restart, you should see in the API response:

```json
{
  "dkp_id": "12017236",
  "columns": ["time", "price", "seller_id", "is_buy_box", "variant_id"],
  "data": [...],
  "variants": [
    {
      "variant_id": "42732793",
      "columns": ["time", "price", "seller_id", "is_buy_box"],
      "data": [[...]]
    },
    {
      "variant_id": "43325143",
      "columns": ["time", "price", "seller_id", "is_buy_box"],
      "data": [[...]]
    }
  ]
}
```

And the chart will show **multiple lines**! 🎉
