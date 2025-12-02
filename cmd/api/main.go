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
	httpAdapter "github.com/mohammadalipour/keepa/internal/adapters/http"
	"github.com/mohammadalipour/keepa/internal/adapters/queue"
	"github.com/mohammadalipour/keepa/internal/adapters/repository"
	"github.com/mohammadalipour/keepa/internal/adapters/scheduler"
	"github.com/mohammadalipour/keepa/internal/core/services"
)

func main() {
	log.Println("Digikala Price Tracker Backend Started...")

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

	log.Println("Connected to database")

	// Initialize repository
	repo := repository.NewPostgresRepository(db)

	// Initialize service
	priceService := services.NewPriceService(repo)

	// Setup HTTP router
	router := httpAdapter.SetupRouter(priceService)

	// Connect to RabbitMQ for scheduler
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	mq, err := queue.NewRabbitMQ(rabbitmqURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v (scheduler disabled)", err)
	} else {
		defer mq.Close()
		log.Println("Connected to RabbitMQ")

		// Declare queue
		if err := mq.DeclareQueue("scrape_tasks"); err != nil {
			log.Printf("Warning: Failed to declare queue: %v", err)
		}

		// Start Hot Products Scheduler
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		schedulerInterval := 5 * time.Minute // Check every 5 minutes
		hotProductsScheduler := scheduler.NewHotProductsScheduler(repo, mq, schedulerInterval)
		go hotProductsScheduler.Start(ctx)
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		os.Exit(0)
	}()

	// Start HTTP server
	port := getEnv("PORT", "8080")
	log.Printf("Server listening on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
