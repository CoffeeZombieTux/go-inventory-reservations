package handler

import (
	"fmt"
	"go-inventory-reservations/internal/application"
	"go-inventory-reservations/internal/logger"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AdminHandler handles all admin-related routes, e.g. stock management, etc.
type AdminHandler struct {
	adminStockOrchestrator application.AdminStockOrchestratorInterface
	stockService           service.StockServiceInterface
	logger                 *logger.Logger
}

// NewAdminHandler creates a new AdminHandler instance.
func NewAdminHandler(
	adminStockOrchestrator application.AdminStockOrchestratorInterface,
	stockService service.StockServiceInterface,
	logger *logger.Logger,
) *AdminHandler {
	return &AdminHandler{
		adminStockOrchestrator: adminStockOrchestrator,
		stockService:           stockService,
		logger:                 logger,
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

	stock, err := ah.adminStockOrchestrator.AdjustInventory(ctx, req)
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
	err := ah.adminStockOrchestrator.DeleteStock(ctx, sku)
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

// GetActiveReservationItemsBySku lists active reservation items by SKU with pagination.
func (ah *AdminHandler) GetActiveReservationItemsBySku(ctx *gin.Context) {
	sku := ctx.Param("sku")
	if sku == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "SKU parameter is required",
		})
		return
	}

	limitParam := ctx.DefaultQuery("limit", "")
	offsetParam := ctx.DefaultQuery("offset", "")

	requestedLimit := 0
	requestedOffset := 0

	if limitParam != "" {
		if val, err := strconv.Atoi(limitParam); err == nil {
			requestedLimit = val
		}
	}

	if offsetParam != "" {
		if val, err := strconv.Atoi(offsetParam); err == nil {
			requestedOffset = val
		}
	}

	items, pagination, message, err := ah.adminStockOrchestrator.GetActiveReservationItemsBySku(
		ctx,
		sku,
		requestedLimit,
		requestedOffset,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list active reservation items: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"sku":        sku,
		"items":      items,
		"pagination": pagination,
		"message":    message,
	})
}
