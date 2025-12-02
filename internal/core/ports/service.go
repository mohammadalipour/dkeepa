package ports

import (
	"context"

	"github.com/mohammadalipour/keepa/internal/core/domain"
)

// PriceService defines business logic for price operations.
type PriceService interface {
	// GetProductHistory retrieves price history in optimized format.
	GetProductHistory(ctx context.Context, dkpID string) (*domain.PriceHistoryResponse, error)
	GetProductHistoryByVariant(ctx context.Context, dkpID, variantID string) (*domain.PriceHistoryResponse, error)
	// SaveProductPrice saves or updates product and price log (from extension or scraper)
	SaveProductPrice(ctx context.Context, product *domain.Product, priceLog *domain.PriceLog) error
}
