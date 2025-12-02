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
