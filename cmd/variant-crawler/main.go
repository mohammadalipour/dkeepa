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
				Color   *VariantAttribute `json:"color"`
				Storage *VariantAttribute `json:"storage"`
				Seller  struct {
					Title string `json:"title"`
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

type VariantAttribute struct {
	Title struct {
		Persian string `json:"fa"`
	} `json:"title_fa"`
}

func main() {
	log.Println("🔍 Variant Crawler Started")

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

	// Get all tracked products
	products, err := getTrackedProducts(db)
	if err != nil {
		log.Fatalf("❌ Failed to get tracked products: %v", err)
	}

	log.Printf("📦 Found %d products to crawl for variants", len(products))

	// Create TLS client
	client, err := scraper.NewTLSClient()
	if err != nil {
		log.Fatalf("❌ Failed to create TLS client: %v", err)
	}

	// Crawl variants for each product
	totalVariants := 0
	for i, product := range products {
		log.Printf("[%d/%d] Crawling variants for product %s (%s)...", i+1, len(products), product.DkpID, product.Title)

		variants, err := crawlProductVariants(client, product.DkpID)
		if err != nil {
			log.Printf("   ⚠️  Failed: %v", err)
			continue
		}

		// Save variants to database
		for _, variant := range variants {
			err := saveVariant(db, variant)
			if err != nil {
				log.Printf("   ⚠️  Failed to save variant %d: %v", variant.VariantID, err)
				continue
			}
		}

		totalVariants += len(variants)
		log.Printf("   ✓ Saved %d variants", len(variants))

		// Rate limiting
		time.Sleep(2 * time.Second)
	}

	log.Printf("✅ Total variants discovered: %d", totalVariants)
	log.Println("🎉 Variant Crawler Completed")
}

// getTrackedProducts fetches all products that should be tracked
func getTrackedProducts(db *sqlx.DB) ([]*domain.Product, error) {
	query := `
		SELECT id, dkp_id, title, is_active, category, crawl_priority, is_tracked
		FROM products
		WHERE is_tracked = true AND is_active = true
		ORDER BY crawl_priority DESC, id
	`

	var products []*domain.Product
	err := db.Select(&products, query)
	return products, err
}

// crawlProductVariants fetches all variants for a product
func crawlProductVariants(client *scraper.TLSClient, dkpID string) ([]*domain.ProductVariant, error) {
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

	var variants []*domain.ProductVariant
	for _, v := range productResp.Data.Product.Variants {
		color := ""
		if v.Color != nil {
			color = v.Color.Title.Persian
		}

		storage := ""
		if v.Storage != nil {
			storage = v.Storage.Title.Persian
		}

		variant := &domain.ProductVariant{
			VariantID:    int64(v.ID),
			DkpID:        dkpID,
			VariantTitle: v.Title.Persian,
			Color:        color,
			Storage:      storage,
			IsActive:     v.Status == "marketable",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		variants = append(variants, variant)
	}

	return variants, nil
}

// saveVariant inserts or updates a variant in the database
func saveVariant(db *sqlx.DB, variant *domain.ProductVariant) error {
	query := `
		INSERT INTO product_variants (variant_id, dkp_id, variant_title, color, storage, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (variant_id) 
		DO UPDATE SET 
			variant_title = EXCLUDED.variant_title,
			color = EXCLUDED.color,
			storage = EXCLUDED.storage,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
	`

	_, err := db.Exec(query, variant.VariantID, variant.DkpID, variant.VariantTitle,
		variant.Color, variant.Storage, variant.IsActive, variant.CreatedAt, variant.UpdatedAt)

	return err
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
