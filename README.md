# Keepa - Digikala Price Tracker

Complete price tracking system for Digikala products with microservices architecture.

## Architecture

- **Backend API**: Gin framework (Go)
- **Worker**: Scraper with anti-detection (Go)
- **Database**: TimescaleDB (PostgreSQL + time-series)
- **Cache**: Redis
- **Queue**: RabbitMQ
- **Extension**: Chrome Extension (React + TypeScript)

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Node.js 20+ (for extension development)
- Go 1.24+ (for local development)

### 1. Configure Environment
```bash
# Copy environment template
cp .env.example .env

# Edit with your credentials (optional for dev)
nano .env
```

📖 See [Environment Configuration Guide](docs/ENVIRONMENT_CONFIG.md) for details.

### 2. Start Backend Services
```bash
# Start all services
make up

# Check status
docker ps

# View logs
make logs

# Stop services
make down
```

### Build Extension
```bash
cd extension
npm install
npm run build

# Or using Docker
docker build --target export --output dist .
```

### Load Extension in Chrome
1. Open `chrome://extensions/`
2. Enable "Developer mode"
3. Click "Load unpacked"
4. Select `extension/dist` folder

## Services

| Service | Port | Description |
|---------|------|-------------|
| Backend API | 8080 | REST API for price data |
| TimescaleDB | 5432 | Time-series database |
| Redis | 6379 | Caching layer |
| RabbitMQ | 5672 | Message queue |
| RabbitMQ Management | 15672 | Web UI (guest/guest) |

## API Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Get price history
curl "http://localhost:8080/api/v1/products/12345/history"
```

## Project Structure

```
.
├── cmd/
│   ├── api/          # Backend API entry point
│   └── worker/       # Scraper worker entry point
├── internal/
│   ├── adapters/     # Infrastructure adapters
│   │   ├── http/     # HTTP handlers & router
│   │   ├── queue/    # RabbitMQ implementation
│   │   ├── repository/ # Database repository
│   │   ├── scheduler/  # Hot products scheduler
│   │   └── scraper/    # Digikala scraper
│   └── core/
│       ├── domain/   # Domain models
│       ├── ports/    # Interfaces
│       └── services/ # Business logic
├── extension/        # Chrome Extension
│   └── src/
│       ├── background/ # Service worker
│       ├── content/    # Content scripts
│       └── popup/      # Extension popup
├── migrations/       # Database migrations
├── docker-compose.yml
├── Dockerfile        # Backend Dockerfile
├── Dockerfile.worker # Worker Dockerfile
└── Makefile
```

## Development

### Backend
```bash
# Run locally
go run ./cmd/api/main.go

# Build
go build -o bin/server ./cmd/api
```

### Worker
```bash
# Run locally
go run ./cmd/worker/main.go
```

### Extension
```bash
cd extension
npm run dev  # Development mode with hot reload
```

## Features

✅ **Anti-Detection Scraping**: TLS fingerprinting with Chrome impersonation
✅ **Time-Series Database**: TimescaleDB with compression
✅ **Optimized API**: Columnar format (~40% bandwidth reduction)
✅ **Shadow DOM**: Complete CSS isolation
✅ **Interactive Charts**: Recharts visualization
✅ **Production Ready**: Multi-stage Docker builds

## License

MIT
