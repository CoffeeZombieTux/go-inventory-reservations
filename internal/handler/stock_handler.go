package handler

import (
	"go-inventory-reservations/internal/apperror"
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
		writeError(
			ctx,
			http.StatusBadRequest,
			apperror.CodeValidationErrorMessage,
			apperror.CodeValidationErrorCode,
			nil,
		)
		return
	}

	stock, err := sh.stockService.GetStockBySku(ctx, sku)
	if err != nil {
		sh.logger.WithError(err).WithFields(logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"sku":        sku,
		}).Error(logger.LogMessageFailedToGetStock)
		status, code, message, details := mapDomainError(err)
		writeError(ctx, status, message, code, details)
		return
	}

	if stock == nil {
		writeError(ctx, http.StatusNotFound, apperror.CodeNotFoundMessage, apperror.CodeNotFoundCode, nil)
		return
	}

	writeSuccess(ctx, http.StatusOK, "Stock fetched", stock)
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
		sh.logger.WithError(err).WithFields(logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"limit":      requestedLimit,
			"offset":     requestedOffset,
		}).Error(logger.LogMessageFailedToListStocks)
		status, code, msg, details := mapDomainError(err)
		writeError(ctx, status, msg, code, details)
		return
	}

	writeSuccess(ctx, http.StatusOK, message, gin.H{
		"stocks":     stocks,
		"pagination": pagination,
	})
}
