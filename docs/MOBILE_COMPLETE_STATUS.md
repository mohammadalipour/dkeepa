# Complete Mobile Phone Data Collection Status

## 📊 Current Status (December 3, 2025)

### What We Have NOW:
- ✅ **100 products** - From initial v1 crawler (limited)
- ✅ **29 variants** - From 11 products only
- ✅ **29 price logs** - One entry per variant

### What's RUNNING:
- 🔄 **Category Crawler v2** - Fetching ALL mobile phones
  - Progress: Page 35/184 (19%)
  - Expected: ~3,680 total products
  - Time remaining: ~35-40 minutes

---

## 🎯 Complete Collection Plan

### Phase 1: Products ⏳ IN PROGRESS
**Crawler:** `category-crawler-v2`
```bash
./category-crawler-v2 --category=mobile-phone --max=0
```

**Status:** 🔄 Running now (started at 00:15:34)
- **Target:** ALL mobile phone products (~3,680)
- **Current:** Page 35/184
- **Duration:** ~45-50 minutes total
- **Workers:** 3 parallel workers
- **Rate:** ~1.5 seconds per page

**What it does:**
- Fetches product metadata (ID, title, status)
- Saves to `products` table
- Sets category = "mobile-phone"
- Marks for tracking (is_tracked = true)

### Phase 2: Variants ⏸️ PENDING
**Crawler:** `variant-crawler`
```bash
./variant-crawler
```

**Status:** ⏸️ Will run after Phase 1 completes
- **Target:** ALL variants for 3,680 products
- **Expected:** 5,000-8,000 variants (avg 1.5-2 per product)
- **Duration:** ~3-6 hours (2 seconds per product)
- **Workers:** Sequential (no parallel yet)

**What it does:**
- For each product, fetch all variants (colors, storage, etc.)
- Extracts variant properties
- Saves to `product_variants` table
- Identifies active/marketable variants

### Phase 3: Prices ⏸️ PENDING
**Crawler:** `price-tracker`
```bash
./price-tracker
```

**Status:** ⏸️ Will run after Phase 2 completes
- **Target:** Current price for ALL active variants
- **Expected:** 5,000-8,000 price logs
- **Duration:** ~2-4 hours
- **Workers:** Grouped by product (efficient API calls)

**What it does:**
- Fetches current selling price for each variant
- Records seller information
- Saves to `price_history` time-series table
- Creates baseline for price tracking

---

## ⏱️ Timeline Estimate

```
Now (00:15)
    │
    ├─ Phase 1: Products (IN PROGRESS)
    │  └─ Duration: 45-50 minutes
    │  └─ Completion: ~01:05
    │
01:05
    │
    ├─ Phase 2: Variants
    │  └─ Duration: 3-6 hours
    │  └─ Completion: ~04:00-07:00
    │
04:00-07:00
    │
    ├─ Phase 3: Prices
    │  └─ Duration: 2-4 hours
    │  └─ Completion: ~06:00-11:00
    │
06:00-11:00 ✅ COMPLETE
```

**Total Time:** 5.5 - 11 hours (depending on network and API response)

---

## 📦 Expected Final Dataset

### Mobile Phone Products (~3,680)
```sql
SELECT COUNT(*) FROM products WHERE category = 'mobile-phone';
-- Expected: ~3,680
```

**Examples:**
- Samsung Galaxy S25 Ultra
- iPhone 16 Pro Max
- Xiaomi Redmi Note 13
- Huawei P60 Pro
- OnePlus 12
- ...and 3,675+ more

### Product Variants (~5,000-8,000)
```sql
SELECT COUNT(*) FROM product_variants 
WHERE dkp_id IN (SELECT dkp_id FROM products WHERE category = 'mobile-phone')
  AND is_active = true;
-- Expected: ~5,000-8,000
```

**Variant Types:**
- Different colors (مشکی، آبی، سفید، etc.)
- Storage options (128GB, 256GB, 512GB, 1TB)
- RAM configurations (6GB, 8GB, 12GB, 16GB)
- Bundle variations (with charger, without charger)

### Price History (~5,000-8,000 initial entries)
```sql
SELECT COUNT(*) FROM price_history
WHERE dkp_id IN (SELECT dkp_id FROM products WHERE category = 'mobile-phone');
-- Expected after first run: ~5,000-8,000
-- Expected after 1 week (4x daily): ~140,000-224,000
-- Expected after 1 month: ~600,000-960,000
```

**What we'll track:**
- Current selling price
- RRP (recommended retail price)
- Seller information
- Timestamp (for trend analysis)
- Buy box status

---

## 🚀 Automated Script

I've created a complete pipeline script:

**Location:** `/scripts/full-mobile-pipeline.sh`

**Usage:**
```bash
# Run complete pipeline (will take 5-11 hours)
./scripts/full-mobile-pipeline.sh

# Or run in background
nohup ./scripts/full-mobile-pipeline.sh > pipeline.log 2>&1 &

# Monitor progress
tail -f pipeline.log
```

**What it does:**
1. ✅ Fetches ALL mobile phone products
2. ✅ Discovers ALL variants for each product
3. ✅ Records initial prices for all variants
4. ✅ Shows comprehensive statistics
5. ✅ Provides sample data preview

---

## 📊 Monitoring Commands

### Check Products Progress
```bash
# Count mobile phone products
docker exec keepa-timescaledb psql -U postgres -d keepa -c \
  "SELECT COUNT(*) FROM products WHERE category='mobile-phone'"

# View recent additions
docker exec keepa-timescaledb psql -U postgres -d keepa -c \
  "SELECT dkp_id, LEFT(title, 50), last_crawled 
   FROM products 
   WHERE category='mobile-phone' 
   ORDER BY last_crawled DESC LIMIT 10"
```

### Check Variants Progress
```bash
# Count variants
docker exec keepa-timescaledb psql -U postgres -d keepa -c \
  "SELECT COUNT(*) FROM product_variants WHERE is_active=true"

# Products with most variants
docker exec keepa-timescaledb psql -U postgres -d keepa -c \
  "SELECT p.title, COUNT(pv.variant_id) as variants
   FROM products p
   JOIN product_variants pv ON p.dkp_id = pv.dkp_id
   WHERE p.category='mobile-phone' AND pv.is_active=true
   GROUP BY p.title
   ORDER BY variants DESC
   LIMIT 10"
```

### Check Price Tracking Progress
```bash
# Count price logs
docker exec keepa-timescaledb psql -U postgres -d keepa -c \
  "SELECT COUNT(*) FROM price_history 
   WHERE dkp_id IN (SELECT dkp_id FROM products WHERE category='mobile-phone')"

# Recent price updates
docker exec keepa-timescaledb psql -U postgres -d keepa -c \
  "SELECT p.title, ph.price, ph.time
   FROM price_history ph
   JOIN products p ON ph.dkp_id = p.dkp_id
   WHERE p.category='mobile-phone'
   ORDER BY ph.time DESC
   LIMIT 10"
```

### Monitor Crawler Logs
```bash
# Category crawler progress
tail -f crawler-mobile.log

# Check crawler process
ps aux | grep category-crawler-v2

# Estimate completion time
# (184 pages - current page) * 1.5 seconds / 60 = minutes remaining
```

---

## 🔄 After Initial Collection

### Daily Price Tracking (4x per day)
```bash
# Add to crontab
0 0,6,12,18 * * * cd /path/to/keepa && ./price-tracker

# This will track: 5,000-8,000 prices × 4 times = 20,000-32,000 logs/day
```

### Weekly Updates

```bash
# Sunday 2 AM: Update product list (new phones)
0 2 * * 0 cd /path/to/keepa && ./category-crawler-v2 --category=mobile-phone

# Sunday 4 AM: Update variants (new colors/storage)
0 4 * * 0 cd /path/to/keepa && ./variant-crawler
```

---

## 📈 Data Growth Projection

### Storage Requirements

| Timeframe | Products | Variants | Price Logs | DB Size (compressed) |
|-----------|----------|----------|------------|---------------------|
| Initial | 3,680 | ~6,500 | ~6,500 | ~10 MB |
| 1 Week | 3,680 | ~6,500 | ~182,000 | ~50 MB |
| 1 Month | 3,700 | ~6,600 | ~792,000 | ~200 MB |
| 6 Months | 3,800 | ~6,800 | ~4,752,000 | ~1.2 GB |
| 1 Year | 3,900 | ~7,000 | ~9,504,000 | ~2.4 GB |

*Note: TimescaleDB compression can reduce size by 70-90%*

### API Usage

| Activity | Frequency | Daily Calls | Monthly Calls |
|----------|-----------|-------------|---------------|
| Products | Weekly | ~526 | ~3,680 |
| Variants | Weekly | ~526 | ~3,680 |
| Prices | 4x daily | ~14,720 | ~441,600 |
| **TOTAL** | - | **~15,772** | **~448,960** |

---

## ✅ Summary

### Current Status:
- 🔄 **Phase 1 IN PROGRESS:** Fetching all 3,680 mobile phones (35/184 pages done)
- ⏸️ **Phase 2 PENDING:** Will discover ~6,500 variants
- ⏸️ **Phase 3 PENDING:** Will record initial ~6,500 prices

### Completion Time:
- ⏰ **Products:** ~40 minutes remaining (~01:05 completion)
- ⏰ **Variants:** ~3-6 hours after products complete
- ⏰ **Prices:** ~2-4 hours after variants complete
- 🎯 **TOTAL:** ~5-11 hours for complete mobile phone dataset

### What You'll Have:
- ✅ Complete catalog of all Digikala mobile phones
- ✅ All color/storage/RAM variants
- ✅ Baseline prices for time-series tracking
- ✅ Ready for automated daily price monitoring
- ✅ Foundation for price drop alerts & analytics

### Next Steps:
1. ⏳ Wait for current crawler to complete (~40 min)
2. ▶️ Run variant crawler (`./variant-crawler`)
3. ▶️ Run price tracker (`./price-tracker`)
4. 🔄 Set up cron jobs for daily tracking
5. 📊 Build analytics & alerts on top of data

**The system will automatically fetch, store, and track prices for ALL mobile phone products and variants!** 🚀
