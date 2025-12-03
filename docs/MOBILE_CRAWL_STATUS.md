# Mobile Phone Category - Crawl Status Report

## 📊 Current Status

### Products Discovery (Category Crawler)
✅ **Status:** COMPLETED  
📦 **Products Discovered:** 1,854 out of 2,000 (fetched with --max=2000)  
📄 **Pages Processed:** ~93 pages out of 184 total pages  
⏱️ **Duration:** 3 minutes 25 seconds  
⚡ **Performance:** 9.74 products/second  

**Note:** Only ~50% of all mobile phones were fetched due to `--max=2000` limit.

### Variants Discovery (Variant Crawler)
⚠️ **Status:** PARTIAL - Only 11 products processed  
🔢 **Variants Found:** 29 variants from 11 products  
📊 **Coverage:** 0.6% (11 out of 1,854 products)  

### Price Tracking (Price Tracker)
⚠️ **Status:** PARTIAL - Only variants from 11 products  
💰 **Price Records:** 29 (one per variant)  
📊 **Coverage:** 0.6% (11 out of 1,854 products)  

---

## 📈 Detailed Breakdown

### Products Table
```
Total Products:      1,854
Active (Marketable): 281  (15.2%)
Inactive:            1,573 (84.8%)
Tracked:             1,854 (100%)
Category:            mobile-phone
Last Crawl:          2025-12-03 00:17:51
```

### Product Variants Table
```
Total Variants:           29
Products with Variants:   11
Average Variants/Product: 2.6
```

### Price History Table
```
Total Price Records:     29
Products with Prices:    11
Variants with Prices:    29
```

---

## ⚠️ What's Missing

### 1. **Incomplete Product Discovery**
- **Fetched:** 1,854 products (~50%)
- **Missing:** ~1,826 products (remaining 91 pages)
- **Total Available:** ~3,680 mobile phones on Digikala

**Reason:** Crawler was run with `--max=2000` limit

**Solution:**
```bash
# Fetch ALL mobile phones (unlimited)
DB_HOST=localhost ./category-crawler-v2 --category=mobile-phone --max=0
```

### 2. **Missing Variants (99.4% incomplete)**
- **Processed:** 11 products
- **Not Processed:** 1,843 products
- **Estimated Missing Variants:** ~3,500-5,000 variants

**Reason:** Variant crawler was only run on the old 100 products, not the new 1,854

**Solution:**
```bash
# Run variant crawler on ALL mobile phone products
DB_HOST=localhost ./variant-crawler
```

### 3. **Missing Price Data (99.4% incomplete)**
- **Tracked:** 29 variants from 11 products
- **Not Tracked:** All variants from 1,843 products
- **Estimated Missing:** ~3,500-5,000 variants without prices

**Reason:** Price tracker only processes products with variants

**Solution:**
```bash
# Run after variant crawler completes
DB_HOST=localhost ./price-tracker
```

---

## 🚀 Complete Pipeline to Get ALL Data

### Step 1: Fetch ALL Mobile Phone Products (~45-50 minutes)
```bash
# Terminal 1
cd /Users/mohammadalipour/Project/keepa
DB_HOST=localhost ./category-crawler-v2 \
  --category=mobile-phone \
  --max=0 \
  --concurrency=3 \
  --delay=1500 \
  > crawler-mobile-full.log 2>&1 &

# Monitor progress
tail -f crawler-mobile-full.log
```

**Expected Result:**
- ~3,680 products
- 184 pages
- Duration: ~45-50 minutes

### Step 2: Discover ALL Variants (~2-3 hours)
```bash
# Terminal 2 (after Step 1 completes)
DB_HOST=localhost ./variant-crawler \
  > crawler-variants-mobile.log 2>&1 &

# Monitor progress
tail -f crawler-variants-mobile.log
```

**Expected Result:**
- ~5,000-8,000 variants
- ~3,680 products processed
- Duration: ~2-3 hours (2 sec delay per product)

### Step 3: Track Initial Prices (~1-2 hours)
```bash
# Terminal 3 (after Step 2 completes)
DB_HOST=localhost ./price-tracker \
  > tracker-prices-mobile.log 2>&1 &

# Monitor progress
tail -f tracker-prices-mobile.log
```

**Expected Result:**
- ~5,000-8,000 price records
- ~3,680 products
- Duration: ~1-2 hours

### Step 4: Schedule Ongoing Price Tracking

```bash
# Add to crontab for 4x daily price updates
crontab -e

# Add this line:
0 0,6,12,18 * * * cd /path/to/keepa && DB_HOST=localhost ./price-tracker >> /var/log/price-tracker.log 2>&1
```

---

## 📊 Expected Final Results

After completing all 3 steps:

| Metric | Current | After Complete Pipeline |
|--------|---------|-------------------------|
| Products | 1,854 | ~3,680 |
| Active Products | 281 | ~500-800 |
| Variants | 29 | ~5,000-8,000 |
| Price Records | 29 | ~5,000-8,000 (initial) |
| Coverage | 0.6% | 100% |

**Price History Growth:**
- Initial: ~5,000-8,000 records
- After 1 day: ~20,000-32,000 (4 tracking runs)
- After 1 week: ~140,000-224,000
- After 1 month: ~600,000-960,000

---

## 🔍 Current Data Quality

### Active vs Inactive Products

```sql
-- Check product status distribution
SELECT 
    is_active,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 1) as percentage
FROM products 
WHERE category = 'mobile-phone'
GROUP BY is_active;
```

**Result:**
- Active (marketable): 281 (15.2%)
- Inactive: 1,573 (84.8%)

**Analysis:** Most products are out of stock or discontinued, which is normal for mobile phones.

### Products Needing Variants

```sql
-- Products without variants
SELECT COUNT(*) 
FROM products p
LEFT JOIN product_variants pv ON p.dkp_id = pv.dkp_id
WHERE p.category = 'mobile-phone' 
AND pv.variant_id IS NULL;
-- Result: 1,843 products
```

### Products Needing Prices

```sql
-- Active products without any price data
SELECT COUNT(*)
FROM products p
WHERE p.category = 'mobile-phone'
AND p.is_active = true
AND NOT EXISTS (
    SELECT 1 FROM price_history ph 
    WHERE ph.dkp_id = p.dkp_id
);
-- Result: 270 active products without prices
```

---

## 💡 Recommendations

### Immediate Actions (Priority Order)

1. **✅ Run Category Crawler to Fetch ALL Products**
   ```bash
   DB_HOST=localhost ./category-crawler-v2 --category=mobile-phone --max=0
   ```
   - Time: ~45 minutes
   - Impact: Get remaining ~1,826 products
   - Priority: HIGH

2. **⏳ Run Variant Crawler on All Products**
   ```bash
   DB_HOST=localhost ./variant-crawler
   ```
   - Time: ~2-3 hours
   - Impact: Discover 5,000-8,000 variants
   - Priority: HIGH

3. **💰 Run Initial Price Tracking**
   ```bash
   DB_HOST=localhost ./price-tracker
   ```
   - Time: ~1-2 hours
   - Impact: Baseline prices for all variants
   - Priority: MEDIUM

4. **📅 Set Up Scheduled Price Tracking**
   - Configure cron for 4x daily runs
   - Impact: Continuous price monitoring
   - Priority: MEDIUM

### Performance Optimization

For faster processing of all 1,854 products:

```bash
# Category Crawler: Use more workers
./category-crawler-v2 --category=mobile-phone --max=0 \
  --concurrency=5 --delay=1000

# Variant Crawler: Modify to use parallel workers
# (Current version is sequential)

# Price Tracker: Already efficient (groups by product)
```

---

## 📝 Summary

**Current State:**
- ✅ 1,854 products discovered (50% of total)
- ⚠️ Only 11 products have variants
- ⚠️ Only 11 products have price data
- ⚠️ 99.4% of products missing variants and prices

**To Achieve 100% Coverage:**
1. Fetch remaining 1,826 products (~45 min)
2. Discover variants for all 3,680 products (~2-3 hours)
3. Track prices for all variants (~1-2 hours)

**Total Time to Complete:** ~4-5 hours of processing

**After Completion:**
- ✅ 100% product coverage (~3,680 products)
- ✅ Complete variant mapping (~5,000-8,000 variants)
- ✅ Full price tracking baseline
- ✅ Ready for automated 4x daily price updates

---

## 🎯 Next Steps

Run this command to complete the full pipeline:

```bash
#!/bin/bash
# full-mobile-pipeline.sh

echo "Starting full mobile phone pipeline..."

# Step 1: Fetch ALL products
echo "Step 1/3: Fetching all mobile phone products..."
DB_HOST=localhost ./category-crawler-v2 --category=mobile-phone --max=0 \
  > logs/crawler-mobile-full.log 2>&1
echo "✅ Products fetched"

# Step 2: Discover variants
echo "Step 2/3: Discovering variants..."
DB_HOST=localhost ./variant-crawler > logs/crawler-variants.log 2>&1
echo "✅ Variants discovered"

# Step 3: Track prices
echo "Step 3/3: Tracking prices..."
DB_HOST=localhost ./price-tracker > logs/tracker-prices.log 2>&1
echo "✅ Prices tracked"

echo "🎉 Pipeline completed!"
echo "Check database for results:"
echo "  Products: SELECT COUNT(*) FROM products WHERE category='mobile-phone';"
echo "  Variants: SELECT COUNT(*) FROM product_variants;"
echo "  Prices: SELECT COUNT(*) FROM price_history;"
```

Save this script and run:
```bash
chmod +x full-mobile-pipeline.sh
./full-mobile-pipeline.sh
```

**Estimated Total Time:** 4-5 hours (can run in background)
