package http

import (
	"github.com/gin-gonic/gin"
	"github.com/mohammadalipour/keepa/internal/adapters/http/handlers"
	"github.com/mohammadalipour/keepa/internal/core/ports"
)

// SetupRouter creates and configures the Gin router.
func SetupRouter(priceService ports.PriceService) *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		priceHandler := handlers.NewPriceHandler(priceService)
		v1.GET("/products/:dkp_id/history", priceHandler.GetProductHistory)
		v1.POST("/products/ingest", priceHandler.IngestProductData)
	}

	return router
}
