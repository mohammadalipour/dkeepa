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

	fmt.Println("🔍 Checking price_history table structure...")
	fmt.Println("")

	type Column struct {
		ColumnName string `db:"column_name"`
		DataType   string `db:"data_type"`
	}

	var columns []Column
	err = db.Select(&columns, `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'price_history'
		ORDER BY ordinal_position
	`)

	if err != nil {
		log.Fatalf("❌ Failed to get columns: %v", err)
	}

	fmt.Println("📋 Columns in price_history table:")
	fmt.Println("-----------------------------------")
	for i, col := range columns {
		fmt.Printf("%d. %-20s %s\n", i+1, col.ColumnName, col.DataType)
	}
	fmt.Println("")

	// Check if variant_id exists
	hasVariantID := false
	for _, col := range columns {
		if col.ColumnName == "variant_id" {
			hasVariantID = true
			break
		}
	}

	if hasVariantID {
		fmt.Println("✅ variant_id column exists")
	} else {
		fmt.Println("❌ variant_id column is MISSING!")
		fmt.Println("")
		fmt.Println("To fix, run migration:")
		fmt.Println("  psql -h localhost -U postgres -d keepa -f migrations/002_add_variant_id.sql")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
