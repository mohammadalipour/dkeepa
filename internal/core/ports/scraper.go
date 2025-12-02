package ports

import (
	"context"

	"github.com/mohammadalipour/keepa/internal/core/domain"
)

// Scraper defines the interface for fetching product data.
type Scraper interface {
	// ScrapeProduct fetches and parses product data for a given DKP ID.
	ScrapeProduct(ctx context.Context, dkpID, variantID string) (*domain.Product, *domain.PriceLog, error)
}
