# Keepa Chrome Extension

Digikala Price Tracker Chrome Extension built with React 18, TypeScript, and Vite.

## Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```

## Loading the Extension

1. Run `npm run dev` or `npm run build`
2. Open Chrome and navigate to `chrome://extensions/`
3. Enable "Developer mode"
4. Click "Load unpacked"
5. Select the `dist` folder

## Features

- **Price History Widget**: Displays on Digikala product pages
- **Background Service Worker**: Fetches data from backend API
- **Popup**: Shows extension status and info

## Architecture

- **Manifest V3**: Modern Chrome extension format
- **React 18**: UI components
- **TypeScript**: Type safety
- **Vite + @crxjs/vite-plugin**: Fast development and building

## API Integration

The extension communicates with the backend API at `http://localhost:8080`.

Endpoint: `GET /api/v1/products/:dkp_id/history`
