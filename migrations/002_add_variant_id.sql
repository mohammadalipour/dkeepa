-- Add variant_id to price_history table
ALTER TABLE price_history ADD COLUMN variant_id TEXT;

-- Update the unique index to include variant_id
-- We need to drop the old chunk_time_interval based index if it exists implicitly or explicitly?
-- TimescaleDB hypertables are partitioned by time.
-- We likely want an index on (dkp_id, variant_id, time) for fast lookups.

CREATE INDEX idx_price_history_dkp_variant_time ON price_history (dkp_id, variant_id, time DESC);
