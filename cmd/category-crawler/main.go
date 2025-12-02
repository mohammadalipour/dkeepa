package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mohammadalipour/keepa/internal/adapters/scraper"
	"github.com/mohammadalipour/keepa/internal/core/domain"
)

const (
	MaxProducts  = 100
	CategorySlug = "mobile-phone"
)

// DigikalaSearchResponse represents the API response structure
type DigikalaSearchResponse struct {
	Status int `json:"status"`
	Data   struct {
		Products []struct {
			ID       int    `json:"id"`
			TitleFa  string `json:"title_fa"`
			TitleEn  string `json:"title_en"`
			Status   string `json:"status"`
		} `json:"products"`
		Pager struct {
			CurrentPage int `json:"current_page"`
			TotalPages  int `json:"total_pages"`
		} `json:"pager"`
	} `json:"data"`
}

func main() {
	log.Println("🚀 Category Crawler Started - Mobile Phones")

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

	// Create TLS client for scraping
	client, err := scraper.NewTLSClient()
	if err != nil {
		log.Fatalf("❌ Failed to create TLS client: %v", err)
	}

	// Crawl category
	ctx := context.Background()
	products, err := crawlCategory(ctx, client, CategorySlug, MaxProducts)
	if err != nil {
		log.Fatalf("❌ Failed to crawl category: %v", err)
	}

	log.Printf("✅ Found %d products in category %s", len(products), CategorySlug)

	// Save products to database
	saved := 0
	for _, product := range products {
		err := saveProduct(db, product)
		if err != nil {
			log.Printf("⚠️  Failed to save product %s: %v", product.DkpID, err)
			continue
		}
		saved++
	}

	// Update category crawl timestamp
	updateCategoryCrawlTime(db, CategorySlug, len(products))

	log.Printf("✅ Saved %d/%d products to database", saved, len(products))
	log.Println("🎉 Category Crawler Completed")
}

// crawlCategory fetches products from a Digikala category
func crawlCategory(ctx context.Context, client *scraper.TLSClient, categorySlug string, maxProducts int) ([]*domain.Product, error) {
	var allProducts []*domain.Product
	page := 1

	for len(allProducts) < maxProducts {
		log.Printf("📄 Fetching page %d...", page)

		// Construct API URL with best-selling sort
		apiURL := fmt.Sprintf("https://api.digikala.com/v1/categories/%s/search/?page=%d&sort=22", categorySlug, page)

		// Fetch data
		resp, err := client.Get(apiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}

		// Parse response
		var searchResp DigikalaSearchResponse
		if err := json.Unmarshal([]byte(resp), &searchResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		if searchResp.Status != 200 {
			return nil, fmt.Errorf("API returned non-200 status: %d", searchResp.Status)
		}

		// Extract products
		for _, p := range searchResp.Data.Products {
			if len(allProducts) >= maxProducts {
				break
			}

			product := &domain.Product{
				DkpID:         fmt.Sprintf("%d", p.ID),
				Title:         p.TitleFa,
				IsActive:      p.Status == "marketable",
				Category:      categorySlug,
				CrawlPriority: 5, // Medium priority
				IsTracked:     true,
				LastCrawled:   ptrTime(time.Now()),
			}

			allProducts = append(allProducts, product)
		}

		log.Printf("   ✓ Found %d products (total: %d)", len(searchResp.Data.Products), len(allProducts))

		// Check if we have more pages
		if searchResp.Data.Pager.CurrentPage >= searchResp.Data.Pager.TotalPages {
			break
		}

		// Rate limiting - be nice to Digikala
		time.Sleep(2 * time.Second)
		page++
	}

	return allProducts, nil
}

// saveProduct inserts or updates a product in the database
func saveProduct(db *sqlx.DB, product *domain.Product) error {
	query := `
		INSERT INTO products (dkp_id, title, is_active, category, crawl_priority, is_tracked, last_crawled, last_scraped_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (dkp_id) 
		DO UPDATE SET 
			title = EXCLUDED.title,
			is_active = EXCLUDED.is_active,
			category = EXCLUDED.category,
			crawl_priority = EXCLUDED.crawl_priority,
			is_tracked = EXCLUDED.is_tracked,
			last_crawled = EXCLUDED.last_crawled
	`

	_, err := db.Exec(query, product.DkpID, product.Title, product.IsActive, 
		product.Category, product.CrawlPriority, product.IsTracked, product.LastCrawled, time.Now())
	
	return err
}

// updateCategoryCrawlTime updates the last_crawled timestamp for the category
func updateCategoryCrawlTime(db *sqlx.DB, categorySlug string, productCount int) {
	query := `
		UPDATE categories 
		SET last_crawled = $1, product_count = $2, updated_at = $1
		WHERE category_slug = $3
	`

	_, err := db.Exec(query, time.Now(), productCount, categorySlug)
	if err != nil {
		log.Printf("⚠️  Failed to update category timestamp: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
