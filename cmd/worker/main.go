package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/mohammadalipour/keepa/internal/adapters/queue"
	"github.com/mohammadalipour/keepa/internal/adapters/repository"
	"github.com/mohammadalipour/keepa/internal/adapters/scraper"
)

func main() {
	log.Println("Starting Scraper Worker...")

	// Connect to RabbitMQ with retry logic
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	var mq *queue.RabbitMQ
	var err error

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		mq, err = queue.NewRabbitMQ(rabbitmqURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(time.Duration(2<<uint(i)) * time.Second) // Exponential backoff
	}
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after %d attempts: %v", maxRetries, err)
	}
	defer mq.Close()

	// Declare queue
	if err := mq.DeclareQueue("scrape_tasks"); err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Connect to Database
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "keepa"),
	)

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewPostgresRepository(db)

	// Initialize scraper
	scraperInstance, err := scraper.NewDigikalaScraper()
	if err != nil {
		log.Fatalf("Failed to create scraper: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
	}()

	// Start consuming
	handler := func(task queue.ScrapeTask) error {
		log.Printf("Scraping product: %s", task.DkpID)

		product, priceLog, err := scraperInstance.ScrapeProduct(ctx, task.DkpID, task.VariantID)
		if err != nil {
			return fmt.Errorf("scrape failed: %w", err)
		}

		// Save to database
		if err := repo.UpsertProduct(ctx, product); err != nil {
			return fmt.Errorf("failed to save product: %w", err)
		}

		if err := repo.InsertPriceLog(ctx, priceLog); err != nil {
			return fmt.Errorf("failed to save price log: %w", err)
		}

		log.Printf("Successfully scraped and saved: %s - %s (Price: %d)", task.DkpID, product.Title, priceLog.Price)
		return nil
	}

	if err := mq.Consume(ctx, "scrape_tasks", handler); err != nil {
		log.Fatalf("Consumer error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
