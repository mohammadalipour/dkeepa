package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mohammadalipour/keepa/internal/core/domain"
	"github.com/mohammadalipour/keepa/internal/core/ports"
)

type DigikalaScraper struct {
	client *TLSClient
}

// NewDigikalaScraper creates a new Digikala scraper.
func NewDigikalaScraper() (ports.Scraper, error) {
	client, err := NewTLSClient()
	if err != nil {
		return nil, err
	}
	return &DigikalaScraper{client: client}, nil
}

// ScrapeProduct fetches product data from Digikala API.
func (s *DigikalaScraper) ScrapeProduct(ctx context.Context, dkpID, variantID string) (*domain.Product, *domain.PriceLog, error) {
	url := fmt.Sprintf("https://api.digikala.com/v2/product/%s", dkpID)
	if variantID != "" {
		url += fmt.Sprintf("?variant_id=%s", variantID)
	}

	// Use TLS client to bypass anti-bot
	respBody, err := s.client.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch product API: %w", err)
	}

	// Parse JSON response
	var apiResp struct {
		Status int `json:"status"`
		Data   struct {
			Product struct {
				ID             int    `json:"id"`
				TitleFA        string `json:"title_fa"`
				IsActive       bool   `json:"is_active"`
				DefaultVariant struct {
					ID    int `json:"id"`
					Price struct {
						SellingPrice int64 `json:"selling_price"`
						RRPPrice     int64 `json:"rrp_price"`
					} `json:"price"`
					Seller struct {
						Title string `json:"title"`
					} `json:"seller"`
				} `json:"default_variant"`
			} `json:"product"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(respBody), &apiResp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if apiResp.Status != 200 {
		return nil, nil, fmt.Errorf("API returned status: %d", apiResp.Status)
	}

	product := apiResp.Data.Product
	variant := product.DefaultVariant

	// If variantID was requested but API returned different variant (shouldn't happen if API works), check it.
	// But usually API returns the requested variant as default_variant.

	price := variant.Price.SellingPrice
	if price == 0 {
		price = variant.Price.RRPPrice
	}

	sellerName := variant.Seller.Title
	if sellerName == "" {
		sellerName = "Digikala"
	}

	prod := &domain.Product{
		DkpID:         dkpID,
		Title:         product.TitleFA,
		IsActive:      product.IsActive,
		LastScrapedAt: timePtr(time.Now()),
	}

	priceLog := &domain.PriceLog{
		DkpID:     dkpID,
		VariantID: strconv.Itoa(variant.ID),
		Price:     int(price),
		SellerID:  sellerName,
		IsBuyBox:  true, // API returns the buy box variant
		Time:      time.Now(),
	}

	return prod, priceLog, nil
}

// extractFromJSONLD extracts data from JSON-LD structured data.
func (s *DigikalaScraper) extractFromJSONLD(doc *goquery.Document, dkpID string) (*domain.Product, *domain.PriceLog, error) {
	var jsonLDData map[string]interface{}

	doc.Find("script[type='application/ld+json']").Each(func(i int, sel *goquery.Selection) {
		jsonStr := sel.Text()
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &data); err == nil {
			if data["@type"] == "Product" {
				jsonLDData = data
			}
		}
	})

	if jsonLDData == nil {
		return nil, nil, fmt.Errorf("JSON-LD not found")
	}

	title, _ := jsonLDData["name"].(string)
	if title == "" {
		return nil, nil, fmt.Errorf("title not found in JSON-LD")
	}

	product := &domain.Product{
		DkpID:         dkpID,
		Title:         title,
		IsActive:      true,
		LastScrapedAt: timePtr(time.Now()),
	}

	// Extract price from offers
	var price int64
	if offers, ok := jsonLDData["offers"].(map[string]interface{}); ok {
		if priceStr, ok := offers["price"].(string); ok {
			if p, err := strconv.ParseFloat(priceStr, 64); err == nil {
				price = int64(p * 100) // Convert to minor units
			}
		} else if priceFloat, ok := offers["price"].(float64); ok {
			price = int64(priceFloat * 100)
		}
	}

	priceLog := &domain.PriceLog{
		Time:     time.Now(),
		DkpID:    dkpID,
		Price:    int(price),
		IsBuyBox: true, // Assume main offer is buy box
	}

	return product, priceLog, nil
}

// extractFromDkStatics extracts data from window.dkStatics JavaScript variable.
func (s *DigikalaScraper) extractFromDkStatics(html, dkpID string) (*domain.Product, *domain.PriceLog, error) {
	// Find window.dkStatics = {...}
	re := regexp.MustCompile(`window\.dkStatics\s*=\s*({.*?});`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return nil, nil, fmt.Errorf("dkStatics not found")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(matches[1]), &data); err != nil {
		return nil, nil, fmt.Errorf("failed to parse dkStatics: %w", err)
	}

	// Navigate to product data (structure may vary)
	productData, ok := data["product"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("product data not found in dkStatics")
	}

	title, _ := productData["title"].(string)
	if title == "" {
		return nil, nil, fmt.Errorf("title not found")
	}

	product := &domain.Product{
		DkpID:         dkpID,
		Title:         title,
		IsActive:      true,
		LastScrapedAt: timePtr(time.Now()),
	}

	// Extract price
	var price int64
	if priceData, ok := productData["default_variant"].(map[string]interface{}); ok {
		if priceObj, ok := priceData["price"].(map[string]interface{}); ok {
			if sellingPrice, ok := priceObj["selling_price"].(float64); ok {
				price = int64(sellingPrice)
			}
		}
	}

	priceLog := &domain.PriceLog{
		Time:     time.Now(),
		DkpID:    dkpID,
		Price:    int(price),
		IsBuyBox: true,
	}

	return product, priceLog, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
