# 🔧 Extension Not Loading - ERR_FILE_NOT_FOUND Fix

## Problem
Chrome is looking for old asset files with different hashes:
- ❌ Looking for: `index.tsx-DBeeKQzU.js` (old)
- ✅ Actual file: `index.tsx-DqBpRi0V.js` (new)

This happens because Chrome caches the extension and doesn't always update properly when you click "Reload".

## Solution: Completely Remove & Re-add Extension

### Step 1: Remove Old Extension
1. Go to: `chrome://extensions/`
2. Find: **"Keepa - Digikala Price Tracker"**
3. Click: **"Remove"** button
4. Confirm removal

### Step 2: Close All Digikala Tabs
- Close any open Digikala product pages
- This ensures old content scripts are cleared

### Step 3: Re-add Extension (Fresh)
1. Still on `chrome://extensions/`
2. Make sure **"Developer mode"** is ON (top-right toggle)
3. Click: **"Load unpacked"**
4. Navigate to and select:
   ```
   /Users/mohammadalipour/Project/keepa/extension/dist
   ```
5. Click: **"Select"**

### Step 4: Verify Installation
You should see:
- ✅ Extension appears in the list
- ✅ No error messages
- ✅ Status shows "Enabled"

### Step 5: Test on Digikala
1. Open a **new tab**
2. Visit: https://www.digikala.com/product/dkp-20758981/?variant_id=73675351
3. Open DevTools (F12 or Cmd+Option+I)
4. Check Console tab for:
   ```
   Keepa content script loaded
   Product ID: 20758981 Variant ID: 73675351
   ```

## Why Simple "Reload" Doesn't Work

Chrome's extension reload button has limitations:
- ❌ May not update web_accessible_resources
- ❌ May not clear cached asset references
- ❌ May not reload content scripts in open tabs
- ✅ Complete removal + re-add ensures fresh start

## Quick Rebuild Command

If you make code changes, use this to rebuild:
```bash
cd /Users/mohammadalipour/Project/keepa/extension
./rebuild.sh
```

Or manually:
```bash
cd /Users/mohammadalipour/Project/keepa/extension
rm -rf dist/
npm run build
```

Then **always remove and re-add** the extension in Chrome.

## Common Mistakes to Avoid

❌ **DON'T**: Just click the reload button on chrome://extensions  
✅ **DO**: Remove and re-add the extension

❌ **DON'T**: Keep old Digikala tabs open  
✅ **DO**: Close all tabs and open fresh ones

❌ **DON'T**: Load from `/extension` folder  
✅ **DO**: Load from `/extension/dist` folder

## Verification Checklist

After re-adding the extension, verify:
- [ ] Extension shows in chrome://extensions with no errors
- [ ] Extension ID changed (new installation)
- [ ] Console shows "Keepa content script loaded" on Digikala
- [ ] No ERR_FILE_NOT_FOUND errors in console
- [ ] Price chart appears on product page

## Still Having Issues?

### Check Service Worker
1. On `chrome://extensions/`, find your extension
2. Click: **"Service worker"** link
3. Check for errors in the DevTools that opens

### Check Content Script
1. Open a Digikala product page
2. Press F12 to open DevTools
3. Go to: Sources tab → Content scripts
4. Verify you see the extension files

### Check Network Tab
1. Open DevTools (F12)
2. Go to: Network tab
3. Refresh the Digikala page
4. Filter by "extension"
5. Look for failed requests (red)

### Hard Refresh Chrome
If all else fails:
1. Remove extension
2. Close **all Chrome windows**
3. Reopen Chrome
4. Re-add extension
