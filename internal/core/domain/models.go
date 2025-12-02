package domain

import (
	"time"
)

// Product represents the metadata of a product.
type Product struct {
	ID            int        `db:"id" json:"id"`
	DkpID         string     `db:"dkp_id" json:"dkp_id"`
	Title         string     `db:"title" json:"title"`
	IsActive      bool       `db:"is_active" json:"is_active"`
	LastScrapedAt *time.Time `db:"last_scraped_at" json:"last_scraped_at"`
	Category      string     `db:"category" json:"category"`
	LastCrawled   *time.Time `db:"last_crawled" json:"last_crawled"`
	CrawlPriority int        `db:"crawl_priority" json:"crawl_priority"`
	IsTracked     bool       `db:"is_tracked" json:"is_tracked"`
}

// ProductVariant represents a variant of a product (different color, storage, etc).
type ProductVariant struct {
	VariantID       int64     `db:"variant_id" json:"variant_id"`
	DkpID           string    `db:"dkp_id" json:"dkp_id"`
	VariantTitle    string    `db:"variant_title" json:"variant_title"`
	Color           string    `db:"color" json:"color"`
	Storage         string    `db:"storage" json:"storage"`
	OtherProperties string    `db:"other_properties" json:"other_properties"` // JSONB stored as string
	IsActive        bool      `db:"is_active" json:"is_active"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// Category represents a Digikala category being tracked.
type Category struct {
	ID           int        `db:"id" json:"id"`
	CategorySlug string     `db:"category_slug" json:"category_slug"`
	CategoryName string     `db:"category_name" json:"category_name"`
	CategoryURL  string     `db:"category_url" json:"category_url"`
	LastCrawled  *time.Time `db:"last_crawled" json:"last_crawled"`
	ProductCount int        `db:"product_count" json:"product_count"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// PriceLog represents a single price point in time.
type PriceLog struct {
	Time      time.Time `db:"time" json:"time"`
	DkpID     string    `db:"dkp_id" json:"dkp_id"`
	VariantID string    `db:"variant_id" json:"variant_id"`
	Price     int       `db:"price" json:"price"`
	SellerID  string    `db:"seller_id" json:"seller_id"`
	IsBuyBox  bool      `db:"is_buy_box" json:"is_buy_box"`
}

// PriceHistoryResponse represents the optimized columnar response format.
type PriceHistoryResponse struct {
	DkpID   string          `json:"dkp_id"`
	Columns []string        `json:"columns"`
	Data    [][]interface{} `json:"data"`
}
