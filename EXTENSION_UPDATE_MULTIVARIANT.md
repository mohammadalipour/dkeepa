# Extension Update - Multi-Variant Chart Support

## 🎯 Changes Made

The Chrome extension has been updated to display **separate lines for each product variant** instead of showing all variants in a single line.

## 📝 What Was Updated

### 1. Interface Updates (`PriceChart.tsx`)

Added new interface to support variant-specific data:
```typescript
interface VariantSeries {
    variant_id: string;
    columns: string[];
    data: any[][];
}

interface PriceData {
    dkp_id: string;
    columns: string[];
    data: any[][];
    variants?: VariantSeries[];  // NEW: Array of variant series
}

interface ChartDataPoint {
    time: number;
    date: string;
    price: number;
    seller_id: string;
    is_buy_box: boolean;
    variant_id?: string;  // NEW: Variant identifier
}
```

### 2. Data Transformation Logic

Updated `fetchPriceHistory()` function to:
- Check if backend returns `variants` array (new format)
- Merge all variant data into a unified timeline
- Add `variant_id` to each data point
- Fall back to old format for backward compatibility

### 3. Chart Rendering

Updated the chart component to:
- Detect unique variant IDs in the data
- Create a separate `<Line>` component for each variant
- Assign unique colors (25-color palette)
- Show variant info in tooltips
- Display legend with variant IDs

### 4. Tooltip Enhancement

Added variant information to tooltips:
```
🏷️ Variant: 44529292
قیمت: 114,875,000 تومان
فروشنده: 1
✓ Buy Box
```

## 🎨 Color Palette

The chart now supports up to 25 different variants with distinct colors:
- Green, Blue, Red, Purple, Orange, Cyan, Pink, etc.
- Colors cycle if more than 25 variants exist

## 🔄 How to Reload the Extension

### Option 1: Reload via Chrome Extensions Page
1. Open Chrome and go to: `chrome://extensions/`
2. Find "Keepa Extension"
3. Click the **🔄 Reload** button

### Option 2: Reload Specific Extension
```bash
# The extension ID might be different on your system
# Check chrome://extensions/ to find your extension ID
```

### Option 3: Complete Reinstall (if needed)
1. Go to `chrome://extensions/`
2. Remove the old extension
3. Click "Load unpacked"
4. Select: `/Users/mohammadalipour/Project/keepa/extension/dist`

## 🧪 Testing

### Test Product (21 Variants)
Visit: https://www.digikala.com/product/dkp-12017236/

**Expected Result:**
- 21 different colored lines
- Each line represents a different variant
- Legend shows "Variant XXXXXXXX" labels
- Tooltip shows variant ID when hovering

### Verification Steps
1. Open the test product page
2. Wait for the chart to load
3. Verify you see multiple colored lines (not just one)
4. Hover over different lines to see variant IDs
5. Check that each color represents a different variant

## 📊 API Response Format

The backend now returns data in this format:

```json
{
  "dkp_id": "12017236",
  "columns": ["time", "price", "seller_id", "is_buy_box", "variant_id"],
  "data": [...],  // Backward compatible flat format
  "variants": [   // NEW: Separate series per variant
    {
      "variant_id": "44529292",
      "columns": ["time", "price", "seller_id", "is_buy_box"],
      "data": [[1764720848, 114875000, "1", true]]
    },
    {
      "variant_id": "72665975",
      "columns": ["time", "price", "seller_id", "is_buy_box"],
      "data": [[1764720848, 113450000, "1", true]]
    }
    // ... more variants
  ]
}
```

## ✅ Build Output

```
✓ 841 modules transformed.
dist/assets/index.tsx-CVK4HMEf.js  469.31 kB │ gzip: 133.83 kB
✓ built in 2.85s
```

## 🐛 Troubleshooting

### Issue: Still seeing single line
**Solution:** Make sure you reloaded the extension after building

### Issue: Colors not distinct
**Solution:** This is expected if variants have very similar price points

### Issue: Legend too long
**Solution:** Variants are shown as "Variant XXXXXXXX" - this is the variant_id from the database

### Issue: Chart performance slow
**Solution:** This is expected for products with 20+ variants. The chart is rendering many data points.

## 📁 Files Modified

1. `extension/src/content/PriceChart.tsx` - Main chart component
   - Updated interfaces
   - Modified data fetching logic
   - Enhanced chart rendering with multiple lines
   - Added variant info to tooltips

2. `extension/dist/*` - Built files (ready to load in Chrome)

## 🚀 Next Steps

After reloading the extension:
1. Test with product 12017236 (21 variants)
2. Test with a product that has fewer variants
3. Verify tooltips show variant IDs correctly
4. Check that legend is readable and functional

## 📸 Expected vs Before

**BEFORE:** Single green line with multiple points on the same day
**AFTER:** Multiple colored lines, one per variant, spread across different days

---

Built: $(date)
Status: ✅ Ready for testing
