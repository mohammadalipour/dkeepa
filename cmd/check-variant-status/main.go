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

	// Check the Samsung Galaxy A06 variants (product 18344402)
	type Variant struct {
		VariantID int64  `db:"variant_id"`
		DkpID     string `db:"dkp_id"`
		IsActive  bool   `db:"is_active"`
	}

	var variants []Variant
	err = db.Select(&variants, `
		SELECT variant_id, dkp_id, is_active
		FROM product_variants
		WHERE dkp_id = '18344402'
		ORDER BY variant_id
	`)

	if err != nil {
		log.Fatalf("❌ Failed to query variants: %v", err)
	}

	fmt.Println("🔍 Samsung Galaxy A06 (DKP: 18344402) - Variants in Database:")
	fmt.Println("-----------------------------------------------------------")
	for _, v := range variants {
		fmt.Printf("Variant ID: %d, IsActive: %v\n", v.VariantID, v.IsActive)
	}
	fmt.Println()
	fmt.Printf("Total: %d variants\n", len(variants))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
