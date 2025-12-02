package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/mohammadalipour/keepa/internal/core/domain"
	"github.com/mohammadalipour/keepa/internal/core/ports"
)

type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new instance of PostgresRepository.
func NewPostgresRepository(db *sqlx.DB) ports.PriceRepository {
	return &PostgresRepository{db: db}
}

// InsertPriceLog inserts a new price record.
func (r *PostgresRepository) InsertPriceLog(ctx context.Context, log *domain.PriceLog) error {
	query := `
		INSERT INTO price_history (time, dkp_id, variant_id, price, seller_id, is_buy_box)
		VALUES (:time, :dkp_id, :variant_id, :price, :seller_id, :is_buy_box)
	`
	_, err := r.db.NamedExecContext(ctx, query, log)
	return err
}

func (r *PostgresRepository) GetProductHistory(ctx context.Context, dkpID string) ([]domain.PriceLog, error) {
	query := `
		SELECT time, dkp_id, variant_id, price, seller_id, is_buy_box
		FROM price_history
		WHERE dkp_id = $1
		ORDER BY time DESC
		LIMIT 1000
	`
	var history []domain.PriceLog
	err := r.db.SelectContext(ctx, &history, query, dkpID)
	return history, err
}

func (r *PostgresRepository) GetProductHistoryByVariant(ctx context.Context, dkpID, variantID string) ([]domain.PriceLog, error) {
	query := `
		SELECT time, dkp_id, variant_id, price, seller_id, is_buy_box
		FROM price_history
		WHERE dkp_id = $1 AND variant_id = $2
		ORDER BY time DESC
		LIMIT 1000
	`
	var history []domain.PriceLog
	err := r.db.SelectContext(ctx, &history, query, dkpID, variantID)
	if err != nil {
		return nil, err
	}
	return history, nil
}

// UpsertProduct creates or updates product metadata.
func (r *PostgresRepository) UpsertProduct(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (dkp_id, title, is_active, last_scraped_at)
		VALUES (:dkp_id, :title, :is_active, :last_scraped_at)
		ON CONFLICT (dkp_id) DO UPDATE SET
			title = EXCLUDED.title,
			is_active = EXCLUDED.is_active,
			last_scraped_at = EXCLUDED.last_scraped_at
	`
	_, err := r.db.NamedExecContext(ctx, query, product)
	return err
}
