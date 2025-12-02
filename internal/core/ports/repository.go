package ports

import (
	"context"

	"github.com/mohammadalipour/keepa/internal/core/domain"
)

// PriceRepository defines the interface for storing and retrieving price data.
type PriceRepository interface {
	// InsertPriceLog inserts a new price record.
	InsertPriceLog(ctx context.Context, log *domain.PriceLog) error

	// GetProductHistory retrieves price history for a product.
	GetProductHistory(ctx context.Context, dkpID string) ([]domain.PriceLog, error)

	// GetProductHistoryByVariant retrieves price history for a specific product variant.
	GetProductHistoryByVariant(ctx context.Context, dkpID, variantID string) ([]domain.PriceLog, error)

	// UpsertProduct creates or updates product metadata.
	UpsertProduct(ctx context.Context, product *domain.Product) error
}
