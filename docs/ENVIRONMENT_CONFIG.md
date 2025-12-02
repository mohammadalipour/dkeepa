# Environment Configuration Guide

This guide explains how to configure the Keepa application using environment variables.

## Quick Start

1. **Copy the example environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit the `.env` file** with your desired configuration:
   ```bash
   nano .env
   # or
   vim .env
   ```

3. **Start the application:**
   ```bash
   docker-compose up -d
   ```

## Environment Variables

### Backend Server

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Backend API server port | `8080` | No |

### Database (TimescaleDB/PostgreSQL)

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | Database host | `timescaledb` | Yes |
| `DB_PORT` | Database port | `5432` | No |
| `DB_USER` | Database username | `postgres` | Yes |
| `DB_PASSWORD` | Database password | `password` | Yes |
| `DB_NAME` | Database name | `keepa` | Yes |

### Redis

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `REDIS_ADDR` | Redis server address | `redis:6379` | Yes |

### RabbitMQ

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `RABBITMQ_URL` | RabbitMQ connection URL | `amqp://guest:guest@rabbitmq:5672/` | Yes |
| `RABBITMQ_DEFAULT_USER` | RabbitMQ username | `guest` | No |
| `RABBITMQ_DEFAULT_PASS` | RabbitMQ password | `guest` | No |

### Scheduler

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SCHEDULER_INTERVAL` | Hot products check interval | `5m` | No |

### Extension

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `BACKEND_API_URL` | Backend API URL for extension | `http://localhost:8080` | Yes |

## Security Best Practices

### Production Environment

**⚠️ IMPORTANT:** Never use default credentials in production!

1. **Change all default passwords:**
   ```env
   DB_PASSWORD=your_strong_password_here
   RABBITMQ_DEFAULT_USER=your_username
   RABBITMQ_DEFAULT_PASS=your_strong_password_here
   ```

2. **Use strong passwords:**
   - Minimum 16 characters
   - Mix of uppercase, lowercase, numbers, and symbols
   - Use a password generator

3. **Restrict database access:**
   ```env
   DB_HOST=your_private_db_host
   ```

4. **Enable SSL/TLS for database:**
   Update the connection string in `cmd/api/main.go`:
   ```go
   dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
       getEnv("DB_USER", "postgres"),
       getEnv("DB_PASSWORD", "password"),
       getEnv("DB_HOST", "localhost"),
       getEnv("DB_PORT", "5432"),
       getEnv("DB_NAME", "keepa"),
   )
   ```

### Development Environment

For development, you can use the default values provided in `.env.example`.

## Docker Compose Integration

The `docker-compose.yml` file is configured to:
1. Load variables from `.env` file automatically
2. Use default values if variables are not set
3. Share environment variables between services

### Verifying Configuration

Check if environment variables are loaded correctly:

```bash
# View backend environment
docker exec keepa-backend env | grep -E "DB_|PORT|REDIS|RABBITMQ"

# View database environment
docker exec keepa-timescaledb env | grep POSTGRES

# View RabbitMQ environment
docker exec keepa-rabbitmq env | grep RABBITMQ
```

## Local Development (without Docker)

If running services locally without Docker:

1. **Set environment variables in your shell:**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=your_password
   export DB_NAME=keepa
   export REDIS_ADDR=localhost:6379
   export RABBITMQ_URL=amqp://guest:guest@localhost:5672/
   export PORT=8080
   ```

2. **Or use a tool like `direnv`:**
   ```bash
   # Install direnv
   brew install direnv  # macOS
   
   # Allow direnv in your project
   echo 'eval "$(direnv hook zsh)"' >> ~/.zshrc
   source ~/.zshrc
   
   # Create .envrc file
   cp .env .envrc
   direnv allow
   ```

3. **Or load from .env file manually:**
   ```bash
   export $(cat .env | xargs)
   ```

## Troubleshooting

### Variables not loading

1. **Check .env file exists:**
   ```bash
   ls -la .env
   ```

2. **Check .env file format:**
   - No spaces around `=`
   - No quotes needed for simple values
   - Use quotes for values with spaces

3. **Restart containers:**
   ```bash
   docker-compose down
   docker-compose up -d
   ```

### Connection issues

1. **Verify service names match:**
   - In Docker: use service name (e.g., `timescaledb`)
   - Locally: use `localhost`

2. **Check port mappings:**
   ```bash
   docker-compose ps
   ```

3. **Test connectivity:**
   ```bash
   # Test database
   docker exec keepa-backend nc -zv timescaledb 5432
   
   # Test RabbitMQ
   docker exec keepa-backend nc -zv rabbitmq 5672
   
   # Test Redis
   docker exec keepa-backend nc -zv redis 6379
   ```

## Example Configurations

### Minimal Production Setup

```env
# Production Environment
PORT=8080
DB_HOST=prod-db.example.com
DB_PORT=5432
DB_USER=keepa_prod
DB_PASSWORD=super_strong_password_12345!@#$%
DB_NAME=keepa_production
REDIS_ADDR=prod-redis.example.com:6379
RABBITMQ_URL=amqp://keepa:secure_password@prod-rabbitmq.example.com:5672/keepa
```

### Development with External Services

```env
# Development with Cloud Services
PORT=8080
DB_HOST=dev-timescale.cloud.com
DB_PORT=5432
DB_USER=dev_user
DB_PASSWORD=dev_password
DB_NAME=keepa_dev
REDIS_ADDR=dev-redis.cloud.com:6379
RABBITMQ_URL=amqp://dev:dev@dev-rabbitmq.cloud.com:5672/
```

## Additional Resources

- [Docker Compose Environment Variables](https://docs.docker.com/compose/environment-variables/)
- [Twelve-Factor App Config](https://12factor.net/config)
- [PostgreSQL Connection Strings](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING)
- [RabbitMQ URI Specification](https://www.rabbitmq.com/uri-spec.html)
