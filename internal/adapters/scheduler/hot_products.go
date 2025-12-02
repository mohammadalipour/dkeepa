package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/mohammadalipour/keepa/internal/adapters/queue"
	"github.com/mohammadalipour/keepa/internal/core/ports"
)

type HotProductsScheduler struct {
	repo     ports.PriceRepository
	mq       *queue.RabbitMQ
	interval time.Duration
}

// NewHotProductsScheduler creates a new scheduler instance.
func NewHotProductsScheduler(repo ports.PriceRepository, mq *queue.RabbitMQ, interval time.Duration) *HotProductsScheduler {
	return &HotProductsScheduler{
		repo:     repo,
		mq:       mq,
		interval: interval,
	}
}

// Start begins the scheduler loop.
func (s *HotProductsScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Printf("Hot Products Scheduler started (interval: %s)", s.interval)

	// Run immediately on start
	s.checkAndQueueHotProducts(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Hot Products Scheduler stopped")
			return
		case <-ticker.C:
			s.checkAndQueueHotProducts(ctx)
		}
	}
}

// checkAndQueueHotProducts identifies products needing updates and queues them.
func (s *HotProductsScheduler) checkAndQueueHotProducts(ctx context.Context) {
	log.Println("Checking for hot products...")

	// TODO: Implement logic to identify hot products
	// For now, we'll use a simple approach:
	// 1. Find products that haven't been scraped in the last hour
	// 2. Queue them for scraping

	// This is a placeholder - in production, you'd query the database
	// for products based on criteria like:
	// - Last scraped time
	// - User interest (view count, watchlist)
	// - Price volatility

	hotProducts := s.getHotProductIDs(ctx)

	for _, dkpID := range hotProducts {
		task := queue.ScrapeTask{DkpID: dkpID}
		if err := s.mq.Publish("scrape_tasks", task); err != nil {
			log.Printf("Failed to queue product %s: %v", dkpID, err)
		} else {
			log.Printf("Queued hot product: %s", dkpID)
		}
	}

	log.Printf("Queued %d hot products", len(hotProducts))
}

// getHotProductIDs returns a list of product IDs that need scraping.
// This is a placeholder implementation.
func (s *HotProductsScheduler) getHotProductIDs(ctx context.Context) []string {
	// TODO: Query database for products that need updates
	// For now, return empty list
	// In production, you would:
	// 1. SELECT DISTINCT dkp_id FROM products WHERE last_scraped_at < NOW() - INTERVAL '1 hour'
	// 2. Or use a more sophisticated algorithm based on user activity

	return []string{}
}
