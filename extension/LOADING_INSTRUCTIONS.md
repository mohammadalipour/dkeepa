# Chrome Extension Loading Instructions

## ✅ Extension Build Complete

The extension has been built successfully and is ready to load!

## 🔧 How to Load the Extension in Chrome

1. **Open Chrome Extensions Page**
   - Navigate to: `chrome://extensions/`
   - Or go to: Menu (⋮) → Extensions → Manage Extensions

2. **Enable Developer Mode**
   - Toggle the "Developer mode" switch in the top-right corner

3. **Load Unpacked Extension**
   - Click "Load unpacked" button
   - Navigate to: `/Users/mohammadalipour/Project/keepa/extension/dist`
   - Click "Select" to load the extension

4. **Verify Installation**
   - You should see "Keepa - Digikala Price Tracker" in your extensions list
   - The extension icon should appear in your Chrome toolbar

## 🧪 How to Test the Extension

1. **Start the Backend** (if not already running)
   ```bash
   cd /Users/mohammadalipour/Project/keepa
   docker-compose up -d backend
   ```

2. **Visit a Digikala Product Page**
   - Example: https://www.digikala.com/product/dkp-20758981/?variant_id=73675351
   - Or any other product page matching the pattern: `https://www.digikala.com/product/dkp-*`

3. **Open Browser Console** (Press F12 or Cmd+Option+I)
   - You should see console logs from the extension:
     - "Keepa content script loaded"
     - "Product ID: XXXXX Variant ID: XXXXX"
     - "Scraping product data..."
     - "✅ Product data sent to backend successfully"

4. **Check Backend Logs**
   ```bash
   docker logs -f keepa-backend
   ```

5. **Verify Data in Database**
   ```bash
   docker exec keepa-timescaledb psql -U postgres -d keepa -c \
     "SELECT time, dkp_id, variant_id, price, seller_id FROM price_history ORDER BY time DESC LIMIT 5;"
   ```

## 🐛 Common Issues & Solutions

### Issue: "Manifest file is missing or unreadable"
**Solution**: Make sure you selected the `/extension/dist` folder, not `/extension`

### Issue: "Icons not loading"
**Solution**: Already fixed! The icon paths have been corrected to use `icons/` instead of `public/icons/`

### Issue: "Extension shows but doesn't inject anything"
**Solution**: 
- Check browser console for errors
- Make sure you're on a Digikala product page (not homepage)
- Try refreshing the page after loading the extension

### Issue: "CORS error when sending to backend"
**Solution**: 
- Make sure backend is running: `docker-compose ps`
- Check host_permissions in manifest includes `http://localhost:8080/*`

### Issue: "No _rch token found"
**Solution**: 
- The scraper tries multiple methods to extract the token
- Check console logs to see which method was attempted
- Some products may not have all data available immediately

## 📊 What the Extension Does

1. **Detects Product Pages**: Automatically activates on Digikala product pages
2. **Extracts Data**: Scrapes product information including:
   - Product ID (dkp_id)
   - Variant ID
   - Title
   - Current price
   - Original price (RRP)
   - Seller name
   - Active status
   - Anti-bot token (_rch)
3. **Sends to Backend**: Forwards data to your Go backend API
4. **Stores in Database**: Backend saves price history in TimescaleDB
5. **Displays Chart** (future): Will show price history chart on the page

## 🔄 After Making Changes

If you modify the extension code:

1. **Rebuild**
   ```bash
   cd /Users/mohammadalipour/Project/keepa/extension
   npm run build
   ```

2. **IMPORTANT: Remove and Re-add Extension**
   - ⚠️ Do NOT just click the reload button!
   - Go to `chrome://extensions/`
   - Click **"Remove"** on the Keepa extension
   - Click **"Load unpacked"** and select `/Users/mohammadalipour/Project/keepa/extension/dist`
   - This ensures all cached files are cleared

**Why?** Chrome's reload button doesn't always update `web_accessible_resources` and may keep old cached asset references, causing `ERR_FILE_NOT_FOUND` errors.

## 📝 Next Steps

After confirming the extension works:
- Monitor price changes over time
- Build out the price chart visualization
- Add popup interface for tracked products
- Implement periodic background checks for hot products
