package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/mohammadalipour/keepa/internal/adapters/scraper"
	"github.com/mohammadalipour/keepa/internal/core/domain"
)

// DigikalaProductResponse represents the product detail API response
type DigikalaProductResponse struct {
	Status int `json:"status"`
	Data   struct {
		Product struct {
			ID       int `json:"id"`
			Variants []struct {
				ID    int `json:"id"`
				Title struct {
					Persian string `json:"fa"`
				} `json:"title_fa"`
				Seller struct {
					Title string `json:"title"`
					ID    int    `json:"id"`
				} `json:"seller"`
				Price struct {
					SellingPrice int `json:"selling_price"`
					RRPPrice     int `json:"rrp_price"`
				} `json:"price"`
				Status string `json:"status"`
			} `json:"variants"`
		} `json:"product"`
	} `json:"data"`
}

func main() {
	log.Println("💰 Price Tracker Started")

	// Connect to database
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "keepa"),
	)

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Connected to database")

	// Get all active variants
	variants, err := getActiveVariants(db)
	if err != nil {
		log.Fatalf("❌ Failed to get variants: %v", err)
	}

	log.Printf("🎯 Found %d variants to track", len(variants))

	// Create TLS client
	client, err := scraper.NewTLSClient()
	if err != nil {
		log.Fatalf("❌ Failed to create TLS client: %v", err)
	}

	// Group variants by product for efficient API calls
	variantsByProduct := make(map[string][]*domain.ProductVariant)
	for _, v := range variants {
		variantsByProduct[v.DkpID] = append(variantsByProduct[v.DkpID], v)
	}

	totalPrices := 0
	productCount := 0

	// Track prices for each product
	for dkpID, productVariants := range variantsByProduct {
		productCount++
		log.Printf("[%d/%d] Tracking prices for product %s (%d variants)...",
			productCount, len(variantsByProduct), dkpID, len(productVariants))

		prices, err := fetchProductPrices(client, dkpID)
		if err != nil {
			log.Printf("   ⚠️  Failed: %v", err)
			continue
		}

		// Save prices to database
		saved := 0
		for _, price := range prices {
			err := savePriceLog(db, price)
			if err != nil {
				log.Printf("   ⚠️  Failed to save price for variant %s: %v", price.VariantID, err)
				continue
			}
			saved++
		}

		totalPrices += saved
		log.Printf("   ✓ Saved %d prices", saved)

		// Rate limiting - be nice to Digikala
		time.Sleep(2 * time.Second)
	}

	log.Printf("✅ Total prices tracked: %d for %d products", totalPrices, productCount)
	log.Println("🎉 Price Tracker Completed")
}

// getActiveVariants fetches all active variants that should be tracked
func getActiveVariants(db *sqlx.DB) ([]*domain.ProductVariant, error) {
	query := `
		SELECT v.variant_id, v.dkp_id, v.variant_title, v.color, v.storage, v.is_active
		FROM product_variants v
		INNER JOIN products p ON v.dkp_id = p.dkp_id
		WHERE v.is_active = true AND p.is_tracked = true AND p.is_active = true
		ORDER BY p.crawl_priority DESC, v.variant_id
	`

	var variants []*domain.ProductVariant
	err := db.Select(&variants, query)
	return variants, err
}

// fetchProductPrices gets current prices for all variants of a product
func fetchProductPrices(client *scraper.TLSClient, dkpID string) ([]*domain.PriceLog, error) {
	apiURL := fmt.Sprintf("https://api.digikala.com/v2/product/%s/", dkpID)

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}

	var productResp DigikalaProductResponse
	if err := json.Unmarshal([]byte(resp), &productResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if productResp.Status != 200 {
		return nil, fmt.Errorf("API returned non-200 status: %d", productResp.Status)
	}

	now := time.Now()
	var priceLogs []*domain.PriceLog

	for _, v := range productResp.Data.Product.Variants {
		if v.Status != "marketable" {
			continue
		}

		priceLog := &domain.PriceLog{
			Time:      now,
			DkpID:     dkpID,
			VariantID: fmt.Sprintf("%d", v.ID),
			Price:     v.Price.SellingPrice,
			SellerID:  fmt.Sprintf("%d", v.Seller.ID),
			IsBuyBox:  true, // Assume buy box for simplicity
		}

		priceLogs = append(priceLogs, priceLog)
	}

	return priceLogs, nil
}

// savePriceLog inserts a price log entry into the database
func savePriceLog(db *sqlx.DB, priceLog *domain.PriceLog) error {
	query := `
		INSERT INTO price_history (time, dkp_id, variant_id, price, seller_id, is_buy_box)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := db.Exec(query, priceLog.Time, priceLog.DkpID, priceLog.VariantID,
		priceLog.Price, priceLog.SellerID, priceLog.IsBuyBox)

	return err
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
