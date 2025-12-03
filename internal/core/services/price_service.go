package services

import (
	"context"

	"github.com/mohammadalipour/keepa/internal/core/domain"
	"github.com/mohammadalipour/keepa/internal/core/ports"
)

type PriceService struct {
	repo ports.PriceRepository
}

// NewPriceService creates a new PriceService instance.
func NewPriceService(repo ports.PriceRepository) ports.PriceService {
	return &PriceService{repo: repo}
}

// GetProductHistory retrieves price history in optimized columnar format.
func (s *PriceService) GetProductHistory(ctx context.Context, dkpID string) (*domain.PriceHistoryResponse, error) {
	logs, err := s.repo.GetProductHistory(ctx, dkpID)
	if err != nil {
		return nil, err
	}
	return s.transformToColumnsWithVariants(dkpID, logs), nil
}

func (s *PriceService) GetProductHistoryByVariant(ctx context.Context, dkpID, variantID string) (*domain.PriceHistoryResponse, error) {
	logs, err := s.repo.GetProductHistoryByVariant(ctx, dkpID, variantID)
	if err != nil {
		return nil, err
	}
	return s.transformToColumnsWithVariants(dkpID, logs), nil
}

func (s *PriceService) transformToColumnsWithVariants(dkpID string, logs []domain.PriceLog) *domain.PriceHistoryResponse {
	response := &domain.PriceHistoryResponse{
		DkpID:    dkpID,
		Columns:  []string{"time", "price", "seller_id", "is_buy_box", "variant_id"},
		Data:     make([][]interface{}, 0, len(logs)),
		Variants: make([]domain.VariantPriceSeries, 0),
	}

	// Group logs by variant
	variantMap := make(map[string][]domain.PriceLog)
	
	for _, log := range logs {
		// Add to combined data (backward compatibility)
		row := []interface{}{
			log.Time.Unix(),
			log.Price,
			log.SellerID,
			log.IsBuyBox,
			log.VariantID,
		}
		response.Data = append(response.Data, row)
		
		// Group by variant
		variantMap[log.VariantID] = append(variantMap[log.VariantID], log)
	}

	// Create separate series for each variant
	for variantID, variantLogs := range variantMap {
		series := domain.VariantPriceSeries{
			VariantID: variantID,
			Columns:   []string{"time", "price", "seller_id", "is_buy_box"},
			Data:      make([][]interface{}, 0, len(variantLogs)),
		}
		
		for _, log := range variantLogs {
			row := []interface{}{
				log.Time.Unix(),
				log.Price,
				log.SellerID,
				log.IsBuyBox,
			}
			series.Data = append(series.Data, row)
		}
		
		response.Variants = append(response.Variants, series)
	}

	return response
}

// SaveProductPrice saves or updates product and adds a price log entry
// Used by both the extension (browser scraping) and worker (automated scraping)
func (s *PriceService) SaveProductPrice(ctx context.Context, product *domain.Product, priceLog *domain.PriceLog) error {
	// First, save or update the product
	if err := s.repo.UpsertProduct(ctx, product); err != nil {
		return err
	}

	// Then, save the price log
	if err := s.repo.InsertPriceLog(ctx, priceLog); err != nil {
		return err
	}

	return nil
}
