-- Migration 003: Add Category Tracking and Variants
-- This migration adds support for category-based product discovery and variant tracking

-- Add category tracking to products table
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS category VARCHAR(100),
ADD COLUMN IF NOT EXISTS last_crawled TIMESTAMP,
ADD COLUMN IF NOT EXISTS crawl_priority INT DEFAULT 1,
ADD COLUMN IF NOT EXISTS is_tracked BOOLEAN DEFAULT true;

-- Create index for faster category queries
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_crawl_priority ON products(crawl_priority DESC);
CREATE INDEX IF NOT EXISTS idx_products_last_crawled ON products(last_crawled);

-- Create product_variants table
CREATE TABLE IF NOT EXISTS product_variants (
    variant_id BIGINT PRIMARY KEY,
    dkp_id VARCHAR(50) NOT NULL,
    variant_title VARCHAR(500),
    color VARCHAR(100),
    storage VARCHAR(100),
    other_properties JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (dkp_id) REFERENCES products(dkp_id) ON DELETE CASCADE
);

-- Create indexes for variants
CREATE INDEX IF NOT EXISTS idx_variants_dkp_id ON product_variants(dkp_id);
CREATE INDEX IF NOT EXISTS idx_variants_is_active ON product_variants(is_active);

-- Create categories table for tracking crawled categories
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    category_slug VARCHAR(100) UNIQUE NOT NULL,
    category_name VARCHAR(200),
    category_url VARCHAR(500),
    last_crawled TIMESTAMP,
    product_count INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create index for categories
CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(category_slug);
CREATE INDEX IF NOT EXISTS idx_categories_is_active ON categories(is_active);

-- Insert initial category for mobile phones
INSERT INTO categories (category_slug, category_name, category_url, is_active)
VALUES ('mobile-phone', 'گوشی موبایل', 'https://www.digikala.com/search/category-mobile-phone/', true)
ON CONFLICT (category_slug) DO NOTHING;

-- Add comment to tables
COMMENT ON TABLE product_variants IS 'Stores all variants (colors, storage options) for each product';
COMMENT ON TABLE categories IS 'Tracks Digikala categories being monitored for price changes';
COMMENT ON COLUMN products.crawl_priority IS 'Higher priority products are crawled more frequently (1-10)';
COMMENT ON COLUMN products.last_crawled IS 'Last time product data was fetched from Digikala';
