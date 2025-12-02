-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Products Table (Relational Metadata)
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    dkp_id VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_scraped_at TIMESTAMPTZ
);

-- Price History Table (Time-Series)
CREATE TABLE IF NOT EXISTS price_history (
    time TIMESTAMPTZ NOT NULL,
    dkp_id VARCHAR(50) NOT NULL,
    price BIGINT NOT NULL,
    seller_id VARCHAR(50),
    is_buy_box BOOLEAN DEFAULT FALSE
);

-- Convert to Hypertable (Partition by time, 7-day chunks)
SELECT create_hypertable('price_history', 'time', chunk_time_interval => INTERVAL '7 days', if_not_exists => TRUE);

-- Enable Compression
ALTER TABLE price_history SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'dkp_id'
);

-- Add Compression Policy (Compress chunks older than 2 days)
SELECT add_compression_policy('price_history', INTERVAL '2 days');
