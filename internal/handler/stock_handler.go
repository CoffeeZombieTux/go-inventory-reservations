package handler

import (
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// StockHandler handles all stock-related routes, e.g. stock retrieval, etc.
type StockHandler struct {
	stockService service.StockServiceInterface
	logger       *logger.Logger
}

// NewStockHandler creates a new StockHandler instance.
func NewStockHandler(stockService service.StockServiceInterface, logger *logger.Logger) *StockHandler {
	return &StockHandler{
		stockService: stockService,
		logger:       logger,
	}
}

// GetStockBySku retrieves a stock item by SKU.
func (sh *StockHandler) GetStockBySku(ctx *gin.Context) {
	sku := ctx.Param("sku")
	if sku == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "SKU parameter is required",
		})
		return
	}

	stock, err := sh.stockService.GetStockBySku(ctx, sku, nil)
	if err != nil {
		sh.logger.Error("Failed to get stock", "error", err.Error(), "sku", sku)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get stock: " + err.Error(),
		})
		return
	}

	if stock == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Stock not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, stock)
}

// GetStocks retrieves a list of stock items, optionally filtered by limit and offset.
func (sh *StockHandler) GetStocks(ctx *gin.Context) {
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

	stocks, pagination, message, err := sh.stockService.GetStocks(ctx, requestedLimit, requestedOffset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list stocks: " + err.Error(),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"stocks":     stocks,
		"pagination": pagination,
		"message":    message,
	})
}
