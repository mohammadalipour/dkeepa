# API Response Format Update

## 🎯 Problem
The current API returns all variants in a single flat array, making it impossible to distinguish which price belongs to which variant (color/storage).

## ✅ Solution
Updated the API response to include **separate price series for each variant**.

## 📊 New Response Format

### Before (Current):
```json
{
  "dkp_id": "2292372",
  "columns": ["time", "price", "seller_id", "is_buy_box"],
  "data": [
    [1733233549, 12990000, "123", true],
    [1733233549, 13490000, "123", true],
    [1733233549, 14990000, "123", true]
  ]
}
```
**Problem**: Can't tell which price is for which variant!

### After (New):
```json
{
  "dkp_id": "2292372",
  "columns": ["time", "price", "seller_id", "is_buy_box", "variant_id"],
  "data": [
    [1733233549, 12990000, "123", true, "14285566"],
    [1733233549, 13490000, "123", true, "42732793"],
    [1733233549, 14990000, "123", true, "43325143"]
  ],
  "variants": [
    {
      "variant_id": "14285566",
      "variant_title": "گوشی موبایل ... - مشکی - 128GB",
      "color": "مشکی",
      "storage": "128GB",
      "columns": ["time", "price", "seller_id", "is_buy_box"],
      "data": [
        [1733233549, 12990000, "123", true],
        [1733233600, 12890000, "123", true]
      ]
    },
    {
      "variant_id": "42732793",
      "variant_title": "گوشی موبایل ... - آبی - 256GB",
      "color": "آبی",
      "storage": "256GB",
      "columns": ["time", "price", "seller_id", "is_buy_box"],
      "data": [
        [1733233549, 13490000, "123", true],
        [1733233600, 13390000, "123", true]
      ]
    }
  ]
}
```

## 🎨 Benefits

1. **Separate Chart Lines**: Each variant can have its own line in the chart
2. **Color/Storage Labels**: Easy to show "Black 128GB" vs "Blue 256GB"
3. **Backward Compatible**: Old `data` field still exists for legacy clients
4. **Efficient**: Single API call returns all variants

## 🔄 Chrome Extension Update Needed

The extension's `PriceChart.tsx` needs to be updated to:

1. **Check for `variants` array** in API response
2. **Create separate series** for each variant
3. **Show variant info** in legend (color + storage)
4. **Different line colors** for each variant

### Example Chart Update:

```typescript
// Before: Single line chart
const chartData = response.data.map(row => ({
  time: row[0],
  price: row[1]
}));

// After: Multi-line chart (one per variant)
const series = response.variants.map(variant => ({
  name: `${variant.color} - ${variant.storage}`,
  data: variant.data.map(row => ({
    time: row[0],
    price: row[1]
  }))
}));
```

## 📋 Implementation Status

- ✅ Backend models updated (`PriceHistoryResponse`, `VariantPriceSeries`)
- ✅ Service layer updated (groups by variant_id)
- ✅ API compiled and ready
- ⏳ Docker backend needs restart to apply changes
- ⏳ Chrome extension needs update to use new format

## 🚀 To Apply Changes

### Restart Backend:
```bash
# Rebuild Docker image
docker-compose build backend

# Restart container
docker-compose up -d backend
```

### Test New API:
```bash
curl "http://localhost:8080/api/v1/products/11346346/history" | jq .variants
```

Should show array of variant price series!

## 📊 Test with Real Data

Product 11346346 (Samsung Galaxy A34) has 1 variant with price history.
Product 12017236 has 21 variants - perfect for testing multi-line charts!

```bash
# Test single variant
curl "http://localhost:8080/api/v1/products/11346346/history"

# Test multiple variants
curl "http://localhost:8080/api/v1/products/12017236/history"
```

Each variant will have its own series in the `variants` array! 🎉
