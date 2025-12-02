#!/bin/bash

# Force Clean Rebuild Script for Chrome Extension

echo "🧹 Cleaning old build artifacts..."
cd "$(dirname "$0")"
rm -rf dist/
rm -rf node_modules/.vite

echo "📦 Rebuilding extension..."
npm run build

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Build successful!"
    echo ""
    echo "🔄 To properly reload in Chrome:"
    echo "1. Go to chrome://extensions/"
    echo "2. Find 'Keepa - Digikala Price Tracker'"
    echo "3. Click 'Remove' button"
    echo "4. Click 'Load unpacked'"
    echo "5. Select: $(pwd)/dist"
    echo ""
    echo "⚠️  IMPORTANT: You MUST remove and re-add the extension!"
    echo "   Simply clicking the reload button may not update all files."
    echo ""
else
    echo "❌ Build failed!"
    exit 1
fi
