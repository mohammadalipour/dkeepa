package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
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

	fmt.Println("🔍 Database Diagnostics")
	fmt.Println("=======================")
	fmt.Println("")

	// Products stats
	var productStats struct {
		Total   int `db:"total"`
		Active  int `db:"active"`
		Tracked int `db:"tracked"`
	}
	err = db.Get(&productStats, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true) as active,
			COUNT(*) FILTER (WHERE is_tracked = true) as tracked
		FROM products
	`)
	if err != nil {
		log.Printf("❌ Failed to get product stats: %v", err)
	} else {
		fmt.Println("📦 Products:")
		fmt.Printf("   Total:   %d\n", productStats.Total)
		fmt.Printf("   Active:  %d\n", productStats.Active)
		fmt.Printf("   Tracked: %d\n", productStats.Tracked)
		fmt.Println("")
	}

	// Variant stats
	var variantStats struct {
		Total  int `db:"total"`
		Active int `db:"active"`
	}
	err = db.Get(&variantStats, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true) as active
		FROM product_variants
	`)
	if err != nil {
		log.Printf("❌ Failed to get variant stats: %v", err)
	} else {
		fmt.Println("🎨 Product Variants:")
		fmt.Printf("   Total:  %d\n", variantStats.Total)
		fmt.Printf("   Active: %d\n", variantStats.Active)
		fmt.Println("")
	}

	// Price stats
	var priceStats struct {
		Total          int    `db:"total"`
		UniqueVariants int    `db:"unique_variants"`
		LatestPrice    string `db:"latest_price"`
	}
	err = db.Get(&priceStats, `
		SELECT 
			COUNT(*) as total,
			COUNT(DISTINCT dkp_id) as unique_variants,
			COALESCE(MAX(time)::text, 'Never') as latest_price
		FROM price_history
	`)
	if err != nil {
		log.Printf("❌ Failed to get price stats: %v", err)
	} else {
		fmt.Println("💰 Price History:")
		fmt.Printf("   Total Entries:     %d\n", priceStats.Total)
		fmt.Printf("   Unique Variants:   %d\n", priceStats.UniqueVariants)
		fmt.Printf("   Latest Price Time: %s\n", priceStats.LatestPrice)

		// Additional price stats
		var recentStats struct {
			Last24h int `db:"last_24h"`
			Last7d  int `db:"last_7d"`
		}
		err = db.Get(&recentStats, `
			SELECT 
				COUNT(*) FILTER (WHERE time > NOW() - INTERVAL '24 hours') as last_24h,
				COUNT(*) FILTER (WHERE time > NOW() - INTERVAL '7 days') as last_7d
			FROM price_history
		`)
		if err == nil {
			fmt.Printf("   Last 24 hours:     %d prices\n", recentStats.Last24h)
			fmt.Printf("   Last 7 days:       %d prices\n", recentStats.Last7d)
		}

		fmt.Println("")
	}

	// Check variants that should be tracked
	var trackableCount int
	err = db.Get(&trackableCount, `
		SELECT COUNT(*)
		FROM product_variants v
		INNER JOIN products p ON v.dkp_id = p.dkp_id
		WHERE v.is_active = true AND p.is_tracked = true AND p.is_active = true
	`)
	if err != nil {
		log.Printf("❌ Failed to get trackable variants: %v", err)
	} else {
		fmt.Println("🎯 Trackable Variants:")
		fmt.Printf("   (is_active AND tracked AND product_active): %d\n", trackableCount)

		// Show price coverage
		if priceStats.UniqueVariants > 0 && trackableCount > 0 {
			coverage := float64(priceStats.UniqueVariants) / float64(trackableCount) * 100
			fmt.Printf("   Price Coverage: %.1f%% (%d/%d variants have prices)\n",
				coverage, priceStats.UniqueVariants, trackableCount)
		}

		fmt.Println("")
	}

	// Sample variants
	type VariantSample struct {
		VariantID      int64  `db:"variant_id"`
		DkpID          string `db:"dkp_id"`
		VariantTitle   string `db:"variant_title"`
		IsActive       bool   `db:"is_active"`
		ProductTracked bool   `db:"product_tracked"`
		ProductActive  bool   `db:"product_active"`
	}

	var samples []VariantSample
	err = db.Select(&samples, `
		SELECT 
			v.variant_id,
			v.dkp_id,
			SUBSTRING(v.variant_title, 1, 50) as variant_title,
			v.is_active,
			p.is_tracked as product_tracked,
			p.is_active as product_active
		FROM product_variants v
		JOIN products p ON v.dkp_id = p.dkp_id
		ORDER BY v.variant_id
		LIMIT 10
	`)
	if err != nil {
		log.Printf("❌ Failed to get sample variants: %v", err)
	} else {
		fmt.Println("🔍 Sample Variants (first 10):")
		fmt.Println("   ID        | DkpID    | Active | Tracked | ProdActive | Title")
		fmt.Println("   " + "----------+----------+--------+---------+------------+-----------------------")
		for _, s := range samples {
			fmt.Printf("   %-9d | %-8s | %-6v | %-7v | %-10v | %s\n",
				s.VariantID, s.DkpID, s.IsActive, s.ProductTracked, s.ProductActive, s.VariantTitle)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
