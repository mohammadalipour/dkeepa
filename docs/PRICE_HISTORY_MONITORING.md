# Price History Monitoring Guide

## Overview
The Keepa system now includes comprehensive monitoring for price history tracking, allowing you to see real-time statistics on price collection coverage and freshness.

## Monitoring Tools

### 1. Quick Database Check
```bash
DB_HOST=localhost ./check-db
```

**Shows:**
- Total price entries
- Unique products/variants with prices
- Latest price timestamp
- Prices from last 24 hours
- Prices from last 7 days
- **Price Coverage**: Percentage of trackable variants that have prices

### 2. Comprehensive System Monitor
```bash
./scripts/monitor-all.sh
```

**Shows:**
- Complete database statistics (products, variants, prices)
- All crawler statuses (running/stopped)
- Progress bars for active crawlers
- Price coverage percentage
- Recommendations for next actions
- Quick command reference

### 3. Continuous Monitoring
```bash
./scripts/monitor-all.sh --loop
```

Refreshes every 10 seconds for real-time monitoring during crawler runs.

### 4. Progress Check (with DB stats)
```bash
./scripts/check-progress.sh
```

Shows crawler progress plus database statistics.

## Key Metrics Explained

### Price Coverage
```
Price Coverage: 1.3% (11/827 variants have prices)
```
- **What it means**: Only 11 out of 827 trackable variants have price data
- **Target**: Should be close to 100% after running price tracker
- **Action**: If below 80%, run `DB_HOST=localhost ./price-tracker`

### Price Freshness
```
Last 24 hours:     29 prices
Last 7 days:       29 prices
Latest Price Time: 2025-12-02 23:03:43
```
- **What it means**: Shows how recent your price data is
- **Target**: Should see new prices every 6 hours (automated)
- **Action**: If no recent prices, run price tracker manually

### Trackable Variants
```
Trackable Variants: 827
(is_active AND tracked AND product_active)
```
- **What it means**: Number of variants that should have prices tracked
- **Calculation**: Only includes active (marketable) variants
- **Note**: This number grows as variant crawler discovers more variants

## Current System Status

Based on latest check:
- ✅ **1,854 products** discovered (281 active/marketable)
- 🔄 **827 active variants** (variant crawler at 61.9% - still running)
- ⚠️ **29 old price entries** (1.3% coverage - needs update)

## Next Steps

1. **Wait for variant crawler to complete** (~3-4 more minutes)
   - Currently at: 174/281 (61.9%)
   - Expected final: ~1,100-1,200 active variants

2. **Run price tracker** once variant crawler completes:
   ```bash
   DB_HOST=localhost ./price-tracker
   ```

3. **Verify results**:
   ```bash
   DB_HOST=localhost ./check-db
   ```
   
   Should show:
   - ✅ ~1,100-1,200 active variants
   - ✅ ~90-100% price coverage
   - ✅ Fresh price data (current timestamp)

## Automated Monitoring

For production, set up cron jobs:

```bash
# Price tracking - every 6 hours
0 */6 * * * cd /path/to/keepa && DB_HOST=localhost ./price-tracker >> /var/log/keepa/prices.log 2>&1

# Variant discovery - weekly (Sundays at 2 AM)
0 2 * * 0 cd /path/to/keepa && DB_HOST=localhost ./variant-crawler >> /var/log/keepa/variants.log 2>&1

# Product discovery - weekly (Sundays at 3 AM)
0 3 * * 0 cd /path/to/keepa && DB_HOST=localhost ./category-crawler-v2 --all >> /var/log/keepa/products.log 2>&1
```

## Troubleshooting

### Low Price Coverage
**Problem**: Price coverage below 80%
**Solution**: Run price tracker manually
```bash
DB_HOST=localhost ./price-tracker
```

### Stale Price Data
**Problem**: No prices in last 24 hours
**Check**: Is price tracker running?
```bash
ps aux | grep price-tracker
```
**Solution**: Run manually if stopped

### Missing Variants
**Problem**: Active variants = 0
**Solution**: Run variant crawler
```bash
DB_HOST=localhost ./variant-crawler
```

## Real-Time Monitoring During Crawls

While crawlers are running:
```bash
# Watch variant crawler progress
tail -f /tmp/variant-crawler-rerun.log

# Watch price tracker progress
tail -f /tmp/price-tracker.log

# Monitor all with auto-refresh
./scripts/monitor-all.sh --loop
```

## API Endpoints (Future)

Consider adding API endpoints for monitoring:
- `GET /api/stats` - Database statistics
- `GET /api/health` - System health check
- `GET /api/coverage` - Price coverage metrics
