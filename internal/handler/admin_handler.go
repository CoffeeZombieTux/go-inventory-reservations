package handler

import (
	"fmt"
	"go-inventory-reservations/internal/logger"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminHandler handles all admin-related routes, e.g. stock management, etc.
type AdminHandler struct {
	stockService service.StockServiceInterface
	logger       *logger.Logger
}

// NewAdminHandler creates a new AdminHandler instance.
func NewAdminHandler(stockService service.StockServiceInterface, logger *logger.Logger) *AdminHandler {
	return &AdminHandler{
		stockService: stockService,
		logger:       logger,
	}
}

// CreateStock creates a new stock item.
func (ah *AdminHandler) CreateStock(ctx *gin.Context) {
	var req apimodel.StockRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	stock, err := ah.stockService.CreateStock(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create stock: " + err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, stock)
}

// UpdateStock updates the stock quantity for a given SKU.
func (ah *AdminHandler) UpdateStock(ctx *gin.Context) {
	var req apimodel.StockRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	stock, err := ah.stockService.AdjustInventory(ctx, req, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update stock qty: " + err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, stock)
}

// DeleteStock deletes a stock item by SKU.
func (ah *AdminHandler) DeleteStock(ctx *gin.Context) {
	sku := ctx.Param("sku")
	if sku == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "SKU parameter is required",
		})
		return
	}
	err := ah.stockService.DeleteStock(ctx, sku)
	if err != nil {
		ah.logger.Error("Failed to get stock", "error", err.Error(), "sku", sku)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete stock: " + err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Stock with SKU '%s' deleted successfully", sku),
		"sku":     sku,
	})
}
