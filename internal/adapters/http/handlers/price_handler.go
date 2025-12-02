package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohammadalipour/keepa/internal/core/domain"
	"github.com/mohammadalipour/keepa/internal/core/ports"
)

type PriceHandler struct {
	service ports.PriceService
}

// NewPriceHandler creates a new PriceHandler.
func NewPriceHandler(service ports.PriceService) *PriceHandler {
	return &PriceHandler{service: service}
}

// GetProductHistory handles GET /api/v1/products/:dkp_id/history
func (h *PriceHandler) GetProductHistory(c *gin.Context) {
	dkpID := c.Param("dkp_id")
	variantID := c.Query("variant_id")

	var response *domain.PriceHistoryResponse
	var err error

	if variantID != "" {
		response, err = h.service.GetProductHistoryByVariant(c.Request.Context(), dkpID, variantID)
	} else {
		response, err = h.service.GetProductHistory(c.Request.Context(), dkpID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// IngestProductData handles POST /api/v1/products/ingest
// Receives product data scraped by the browser extension
func (h *PriceHandler) IngestProductData(c *gin.Context) {
	var req struct {
		DkpID      string `json:"dkp_id" binding:"required"`
		VariantID  string `json:"variant_id"`
		Title      string `json:"title" binding:"required"`
		Price      int    `json:"price" binding:"required"`
		RrpPrice   int    `json:"rrp_price"`
		SellerName string `json:"seller_name"`
		IsActive   bool   `json:"is_active"`
		RchToken   string `json:"rch_token"` // Store for potential future use
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create product and price log from extension data
	now := time.Now()
	product := &domain.Product{
		DkpID:         req.DkpID,
		Title:         req.Title,
		IsActive:      req.IsActive,
		LastScrapedAt: &now,
	}

	priceLog := &domain.PriceLog{
		Time:      time.Now(),
		DkpID:     req.DkpID,
		VariantID: req.VariantID,
		Price:     req.Price,
		SellerID:  req.SellerName,
		IsBuyBox:  true, // Extension scrapes the current buy box
	}

	// Save to database via service
	if err := h.service.SaveProductPrice(c.Request.Context(), product, priceLog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product data ingested successfully",
		"dkp_id":  req.DkpID,
	})
}
