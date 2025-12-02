# Price Format Issue - Fixed

## Problem
The chart was showing incorrect average prices:
- Displayed: **۱٬۶۸۸٬۴۴۵٬۰۰۰** (1.6 billion Rials)
- Some entries had: **۳٬۰۶۹٬۹۰۰٬۰۰۰** (3 billion Rials - 10x too high!)

## Root Cause
Some price entries in the database had an extra zero (multiplied by 10 incorrectly).

## Currency Context
- **Rial**: Iran's official currency (smallest unit)
- **Toman**: Informal unit = 10 Rials
- Digikala displays prices in **Tomans** but their API returns **Rials**

### Example for Lenovo Laptop:
- Display price: **30,699,000 تومان** (30.7 million Tomans)
- API price: **306,990,000 Rials** (306.9 million Rials)
- ✅ Correct: 306,990,000 Rials
- ❌ Wrong: 3,069,900,000 Rials (had extra zero)

## Solution

### 1. Fixed Database
Corrected the inflated prices:
```sql
UPDATE price_history 
SET price = price / 10 
WHERE price > 1000000000 AND dkp_id = '20758981';
```

### 2. Verified Data
```sql
SELECT MIN(price), AVG(price)::bigint, MAX(price) 
FROM price_history 
WHERE dkp_id = '20758981' AND variant_id = '73675351';
```

Result:
```
min        | avg        | max
-----------+------------+-----------
306990000  | 306990000  | 306990000
```

All three entries now show **306,990,000 Rials** ✅

## Expected Chart Display

After refreshing the page, you should see:
- **کمترین (Min)**: ۳۰۶٬۹۹۰٬۰۰۰
- **میانگین (Avg)**: ۳۰۶٬۹۹۰٬۰۰۰
- **بیشترین (Max)**: ۳۰۶٬۹۹۰٬۰۰۰

All three will be the same since all price entries are identical (same product, same time period).

## Price Storage Format

Our system stores prices in **Rials** (the API format):
- 1 Toman = 10 Rials
- **306,990,000 Rials** = **30,699,000 Tomans**

The chart displays in Rials (raw values) with Persian number formatting.

## Why Some Entries Had Extra Zero

The issue may have been:
1. Manual test data insertion with wrong format
2. Early testing before proper API integration
3. DOM extraction fallback multiplying by 10 when it shouldn't

The current API-based scraper (lines 183-184 in digikalaScraper.ts) correctly stores the raw API value without modification:
```typescript
price: defaultVariant.price?.selling_price || defaultVariant.price?.rrp_price || 0,
```

## Testing
1. Refresh the Digikala product page
2. Check the chart - all three stats should be equal
3. Future price changes will show different min/avg/max values

## Note on Chart Display
The chart shows prices in **Rials** (full numbers with Persian digits). If you want to display in Tomans, we can divide by 10 in the display logic, but the database storage should remain in Rials for precision.
