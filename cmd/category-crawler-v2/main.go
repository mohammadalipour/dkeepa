package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mohammadalipour/keepa/internal/adapters/scraper"
	"github.com/mohammadalipour/keepa/internal/core/domain"
)

// Configuration
const (
	DefaultMaxProducts    = 0     // 0 = unlimited (fetch all)
	DefaultConcurrency    = 3     // Parallel requests (be careful with rate limiting)
	DefaultBatchSize      = 50    // Products to save in one batch
	DefaultDelayMs        = 2000  // Milliseconds between requests
	DefaultCrawlPriority  = 5     // Default priority for new products
)

// Available categories
var AvailableCategories = map[string]string{
	"mobile-phone":    "گوشی موبایل",
	"tablet":          "تبلت",
	"laptop":          "لپ‌تاپ",
	"smart-watch":     "ساعت هوشمند",
	"headphone":       "هدفون",
	"keyboard-mouse":  "کیبورد و ماوس",
	"monitor":         "مانیتور",
	"console-gaming":  "کنسول بازی",
	"camera":          "دوربین",
	"speaker":         "اسپیکر",
}

// Command line flags
var (
	categorySlug  = flag.String("category", "mobile-phone", "Category slug to crawl")
	maxProducts   = flag.Int("max", DefaultMaxProducts, "Maximum products to fetch (0 = all)")
	concurrency   = flag.Int("concurrency", DefaultConcurrency, "Number of parallel requests")
	batchSize     = flag.Int("batch", DefaultBatchSize, "Batch size for database inserts")
	delayMs       = flag.Int("delay", DefaultDelayMs, "Delay between requests (ms)")
	dryRun        = flag.Bool("dry-run", false, "Dry run mode (don't save to database)")
	listCategories = flag.Bool("list", false, "List available categories")
	allCategories = flag.Bool("all", false, "Crawl all available categories")
)

// Statistics
type CrawlerStats struct {
	TotalProducts   int64
	SavedProducts   int64
	FailedProducts  int64
	TotalPages      int64
	StartTime       time.Time
	CategoryResults map[string]*CategoryResult
	mu              sync.Mutex
}

type CategoryResult struct {
	CategorySlug string
	ProductCount int64
	Duration     time.Duration
	Error        error
}

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
			ItemCount   int `json:"item_count"`
		} `json:"pager"`
	} `json:"data"`
}

func main() {
	flag.Parse()

	// List categories and exit
	if *listCategories {
		fmt.Println("📋 Available Categories:")
		for slug, name := range AvailableCategories {
			fmt.Printf("  - %s: %s\n", slug, name)
		}
		os.Exit(0)
	}

	log.Println("🚀 High-Performance Category Crawler Started")
	log.Printf("⚙️  Configuration: max=%d, concurrency=%d, batch=%d, delay=%dms", 
		*maxProducts, *concurrency, *batchSize, *delayMs)

	// Connect to database
	db, err := connectDB()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Connected to database")

	// Initialize stats
	stats := &CrawlerStats{
		StartTime:       time.Now(),
		CategoryResults: make(map[string]*CategoryResult),
	}

	// Determine which categories to crawl
	var categoriesToCrawl []string
	if *allCategories {
		for slug := range AvailableCategories {
			categoriesToCrawl = append(categoriesToCrawl, slug)
		}
		log.Printf("🌐 Crawling ALL categories (%d total)", len(categoriesToCrawl))
	} else {
		if _, exists := AvailableCategories[*categorySlug]; !exists {
			log.Fatalf("❌ Unknown category: %s (use --list to see available categories)", *categorySlug)
		}
		categoriesToCrawl = []string{*categorySlug}
		log.Printf("📱 Crawling category: %s (%s)", *categorySlug, AvailableCategories[*categorySlug])
	}

	// Crawl each category
	for _, slug := range categoriesToCrawl {
		categoryName := AvailableCategories[slug]
		log.Printf("\n" + strings.Repeat("=", 60))
		log.Printf("📂 Starting crawl for: %s (%s)", categoryName, slug)
		log.Printf(strings.Repeat("=", 60))

		result := crawlCategoryWithStats(db, slug, stats)
		stats.mu.Lock()
		stats.CategoryResults[slug] = result
		stats.mu.Unlock()

		if result.Error != nil {
			log.Printf("❌ Category %s failed: %v", slug, result.Error)
		} else {
			log.Printf("✅ Category %s completed: %d products in %s", 
				slug, result.ProductCount, result.Duration)
		}
	}

	// Print final statistics
	printFinalStats(stats)
}

func crawlCategoryWithStats(db *sqlx.DB, categorySlug string, stats *CrawlerStats) *CategoryResult {
	startTime := time.Now()
	result := &CategoryResult{
		CategorySlug: categorySlug,
	}

	// Create TLS client
	client, err := scraper.NewTLSClient()
	if err != nil {
		result.Error = fmt.Errorf("failed to create TLS client: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Crawl category
	ctx := context.Background()
	products, err := crawlCategoryParallel(ctx, client, categorySlug, *maxProducts, *concurrency, stats)
	if err != nil {
		result.Error = fmt.Errorf("failed to crawl: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	log.Printf("✅ Found %d products in category %s", len(products), categorySlug)

	// Save to database (in batches)
	if !*dryRun {
		saved, failed := saveProductsBatch(db, products, *batchSize)
		atomic.AddInt64(&stats.SavedProducts, int64(saved))
		atomic.AddInt64(&stats.FailedProducts, int64(failed))
		log.Printf("✅ Saved %d/%d products to database", saved, len(products))

		// Update category metadata
		updateCategoryCrawlTime(db, categorySlug, len(products))
	} else {
		log.Printf("🔄 Dry run mode - skipped database save")
	}

	result.ProductCount = int64(len(products))
	result.Duration = time.Since(startTime)
	return result
}

// crawlCategoryParallel fetches products with parallel page fetching
func crawlCategoryParallel(ctx context.Context, client *scraper.TLSClient, categorySlug string, maxProducts, concurrency int, stats *CrawlerStats) ([]*domain.Product, error) {
	// First, fetch page 1 to get total pages
	firstPage, totalPages, err := fetchPage(client, categorySlug, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch first page: %w", err)
	}

	log.Printf("📊 Total pages: %d", totalPages)
	atomic.AddInt64(&stats.TotalPages, int64(totalPages))

	allProducts := make([]*domain.Product, 0, totalPages*20)
	productsMu := sync.Mutex{}
	
	// Add first page products
	allProducts = append(allProducts, firstPage...)
	atomic.AddInt64(&stats.TotalProducts, int64(len(firstPage)))

	// If unlimited or we need more products, fetch remaining pages
	if maxProducts == 0 || len(allProducts) < maxProducts {
		// Create work queue
		pageQueue := make(chan int, totalPages)
		for page := 2; page <= totalPages; page++ {
			pageQueue <- page
		}
		close(pageQueue)

		// Worker pool
		var wg sync.WaitGroup
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for page := range pageQueue {
					// Check if we've reached max products
					if maxProducts > 0 {
						productsMu.Lock()
						currentCount := len(allProducts)
						productsMu.Unlock()
						
						if currentCount >= maxProducts {
							continue
						}
					}

					log.Printf("👷 Worker %d: Fetching page %d/%d", workerID, page, totalPages)

					products, _, err := fetchPage(client, categorySlug, page)
					if err != nil {
						log.Printf("⚠️  Worker %d: Failed to fetch page %d: %v", workerID, page, err)
						continue
					}

					productsMu.Lock()
					allProducts = append(allProducts, products...)
					atomic.AddInt64(&stats.TotalProducts, int64(len(products)))
					productsMu.Unlock()

					// Rate limiting
					time.Sleep(time.Duration(*delayMs) * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()
	}

	// Trim to max if specified
	if maxProducts > 0 && len(allProducts) > maxProducts {
		allProducts = allProducts[:maxProducts]
	}

	return allProducts, nil
}

// fetchPage fetches a single page of products
func fetchPage(client *scraper.TLSClient, categorySlug string, page int) ([]*domain.Product, int, error) {
	apiURL := fmt.Sprintf("https://api.digikala.com/v1/categories/%s/search/?page=%d&sort=22", categorySlug, page)

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, 0, err
	}

	var searchResp DigikalaSearchResponse
	if err := json.Unmarshal([]byte(resp), &searchResp); err != nil {
		return nil, 0, err
	}

	if searchResp.Status != 200 {
		return nil, 0, fmt.Errorf("API returned status: %d", searchResp.Status)
	}

	products := make([]*domain.Product, 0, len(searchResp.Data.Products))
	for _, p := range searchResp.Data.Products {
		product := &domain.Product{
			DkpID:         fmt.Sprintf("%d", p.ID),
			Title:         p.TitleFa,
			IsActive:      p.Status == "marketable",
			Category:      categorySlug,
			CrawlPriority: DefaultCrawlPriority,
			IsTracked:     true,
			LastCrawled:   ptrTime(time.Now()),
		}
		products = append(products, product)
	}

	return products, searchResp.Data.Pager.TotalPages, nil
}

// saveProductsBatch saves products in batches for better performance
func saveProductsBatch(db *sqlx.DB, products []*domain.Product, batchSize int) (saved, failed int) {
	for i := 0; i < len(products); i += batchSize {
		end := i + batchSize
		if end > len(products) {
			end = len(products)
		}

		batch := products[i:end]
		log.Printf("💾 Saving batch %d-%d of %d", i+1, end, len(products))

		for _, product := range batch {
			err := saveProduct(db, product)
			if err != nil {
				log.Printf("⚠️  Failed to save product %s: %v", product.DkpID, err)
				failed++
				continue
			}
			saved++
		}
	}

	return saved, failed
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
			last_crawled = EXCLUDED.last_crawled
	`

	_, err := db.Exec(query, product.DkpID, product.Title, product.IsActive, 
		product.Category, product.CrawlPriority, product.IsTracked, product.LastCrawled, time.Now())
	
	return err
}

// updateCategoryCrawlTime updates the last_crawled timestamp for the category
func updateCategoryCrawlTime(db *sqlx.DB, categorySlug string, productCount int) {
	query := `
		INSERT INTO categories (category_slug, category_name, category_url, last_crawled, product_count, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		ON CONFLICT (category_slug) 
		DO UPDATE SET 
			last_crawled = EXCLUDED.last_crawled,
			product_count = EXCLUDED.product_count,
			updated_at = NOW()
	`

	categoryName := AvailableCategories[categorySlug]
	categoryURL := fmt.Sprintf("https://www.digikala.com/search/category-%s/", categorySlug)

	_, err := db.Exec(query, categorySlug, categoryName, categoryURL, time.Now(), productCount)
	if err != nil {
		log.Printf("⚠️  Failed to update category timestamp: %v", err)
	}
}

func printFinalStats(stats *CrawlerStats) {
	duration := time.Since(stats.StartTime)
	
	log.Printf("\n" + strings.Repeat("=", 60))
	log.Println("📊 FINAL STATISTICS")
	log.Printf(strings.Repeat("=", 60))
	log.Printf("⏱️  Total Duration: %s", duration)
	log.Printf("📦 Total Products: %d", stats.TotalProducts)
	log.Printf("💾 Saved Products: %d", stats.SavedProducts)
	log.Printf("❌ Failed Products: %d", stats.FailedProducts)
	log.Printf("📄 Total Pages: %d", stats.TotalPages)
	
	if stats.TotalProducts > 0 {
		avgTime := duration.Seconds() / float64(stats.TotalProducts)
		log.Printf("⚡ Average: %.2f products/second (%.3fs per product)", 
			float64(stats.TotalProducts)/duration.Seconds(), avgTime)
	}

	log.Println("\n📂 Per-Category Results:")
	for slug, result := range stats.CategoryResults {
		status := "✅"
		if result.Error != nil {
			status = "❌"
		}
		log.Printf("  %s %s (%s): %d products in %s", 
			status, AvailableCategories[slug], slug, result.ProductCount, result.Duration)
	}
	
	log.Printf(strings.Repeat("=", 60))
	log.Println("🎉 Crawler Completed!")
}

func connectDB() (*sqlx.DB, error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "keepa"),
	)

	return sqlx.Connect("pgx", dbURL)
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
