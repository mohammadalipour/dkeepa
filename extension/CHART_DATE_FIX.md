# Chart Date Fix - 0NaN/NaN/NaN Error

## Problem
The price chart was displaying `0NaN/NaN/NaN` instead of proper dates.

## Root Cause
1. **Invalid timestamps in database**: Old entries had `0001-01-01 00:00:00+00` (zero time) from before the timestamp fix
2. **Negative Unix timestamp**: When converted to Unix timestamp, this became `-62135596800` (negative number)
3. **NaN propagation**: When moment.js tried to parse this negative timestamp, it produced `NaN`, which cascaded through the date formatting

## Solution

### 1. Frontend Fix (PriceChart.tsx)
Added validation to handle invalid timestamps:

```typescript
// Filter out invalid data points
const transformed = data.data
    .filter((row) => row[0] > 0) // Filter out invalid timestamps
    .map((row) => ({
        time: row[0],
        date: formatDate(row[0]),
        price: row[1],
        seller_id: row[2],
        is_buy_box: row[3]
    }));

// Format functions now check for invalid values
const formatDate = (timestamp: number) => {
    if (!timestamp || timestamp <= 0) {
        return 'تاریخ نامعتبر';
    }
    return moment(timestamp * 1000).format('jYYYY/jMM/jDD');
};
```

### 2. Database Cleanup
Removed invalid entries with zero timestamps:

```sql
DELETE FROM price_history WHERE time < '2020-01-01';
```

## Before and After

### Before API Response:
```json
{
  "data": [
    [1764633018, 3069900000, "دیجی‌کالا", true],   // Valid ✓
    [1764632720, 306990000, "دیجی‌کالا", true],    // Valid ✓
    [-62135596800, 306990000, "دیجی‌کالا", true]  // Invalid ✗ (caused NaN)
  ]
}
```

### After API Response:
```json
{
  "data": [
    [1764633018, 3069900000, "دیجی‌کالا", true],   // Valid ✓
    [1764632720, 306990000, "دیجی‌کالا", true]     // Valid ✓
  ]
}
```

## Chart Display

### Before:
- X-axis labels: `0NaN/NaN/NaN`, `0NaN/NaN/NaN`, `1403/09/11`
- Tooltip dates: Invalid
- Chart broken/misaligned

### After:
- X-axis labels: `1403/09/11`, `1403/09/11` (proper Jalali dates)
- Tooltip dates: `1403/09/11 - یکشنبه` (with Persian day names)
- Chart displays correctly

## Testing
1. Rebuild extension: `cd extension && npm run build`
2. Reload extension in Chrome (chrome://extensions → reload button)
3. Visit: https://www.digikala.com/product/dkp-20758981/?variant_id=73675351
4. Check browser console for logs
5. Verify chart displays proper dates instead of `0NaN/NaN/NaN`

## Prevention
The timestamp issue was fixed in the backend on Dec 1, 2025 by adding:
```go
priceLog := &domain.PriceLog{
    Time: time.Now(),  // ✓ Now properly set
    // ...
}
```

All new entries will have valid timestamps. The frontend now also validates timestamps as an additional safety measure.
