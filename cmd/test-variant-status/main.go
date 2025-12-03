package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mohammadalipour/keepa/internal/adapters/scraper"
)

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
				Status string `json:"status"`
			} `json:"variants"`
		} `json:"product"`
	} `json:"data"`
}

func main() {
	// Test with a known product (one that has multiple variants)
	dkpID := "18344402" // Samsung Galaxy A06 - should have 8 variants

	client, err := scraper.NewTLSClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	apiURL := fmt.Sprintf("https://api.digikala.com/v2/product/%s/", dkpID)
	fmt.Printf("🔍 Fetching: %s\n\n", apiURL)

	resp, err := client.Get(apiURL)
	if err != nil {
		log.Fatalf("Failed to fetch: %v", err)
	}

	var productResp DigikalaProductResponse
	if err := json.Unmarshal([]byte(resp), &productResp); err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	fmt.Printf("📦 Product ID: %d\n", productResp.Data.Product.ID)
	fmt.Printf("🎨 Found %d variants:\n\n", len(productResp.Data.Product.Variants))

	for i, v := range productResp.Data.Product.Variants {
		fmt.Printf("[%d] Variant ID: %d\n", i+1, v.ID)
		fmt.Printf("    Title: %s\n", v.Title.Persian)
		fmt.Printf("    Status: '%s'\n", v.Status)
		fmt.Printf("    Is Marketable: %v\n", v.Status == "marketable")
		fmt.Println()
	}
}
