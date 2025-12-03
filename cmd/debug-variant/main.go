package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mohammadalipour/keepa/internal/adapters/scraper"
	"github.com/mohammadalipour/keepa/internal/core/domain"
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
				Color   *VariantAttribute `json:"color"`
				Storage *VariantAttribute `json:"storage"`
				Status  string            `json:"status"`
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
	dkpID := "18344402"

	client, err := scraper.NewTLSClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	apiURL := fmt.Sprintf("https://api.digikala.com/v2/product/%s/", dkpID)
	fmt.Printf("🔍 Fetching: %s\n", apiURL)

	resp, err := client.Get(apiURL)
	if err != nil {
		log.Fatalf("Failed to fetch: %v", err)
	}

	var productResp DigikalaProductResponse
	if err := json.Unmarshal([]byte(resp), &productResp); err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	fmt.Printf("\n📦 Processing %d variants...\n\n", len(productResp.Data.Product.Variants))

	var variants []*domain.ProductVariant
	for i, v := range productResp.Data.Product.Variants {
		color := ""
		if v.Color != nil {
			color = v.Color.Title.Persian
		}

		storage := ""
		if v.Storage != nil {
			storage = v.Storage.Title.Persian
		}

		isActive := v.Status == "marketable"

		fmt.Printf("[%d] Variant ID: %d\n", i+1, v.ID)
		fmt.Printf("    API Status: '%s'\n", v.Status)
		fmt.Printf("    Comparison: '%s' == 'marketable' = %v\n", v.Status, isActive)
		fmt.Printf("    Will set IsActive: %v\n\n", isActive)

		variant := &domain.ProductVariant{
			VariantID:    int64(v.ID),
			DkpID:        dkpID,
			VariantTitle: v.Title.Persian,
			Color:        color,
			Storage:      storage,
			IsActive:     isActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		variants = append(variants, variant)
	}

	fmt.Println("✅ All variants processed successfully")
	fmt.Printf("📊 Summary: %d variants, all should have IsActive=true\n", len(variants))
}
