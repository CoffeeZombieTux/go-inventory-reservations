package handler

import (
	"go-inventory-reservations/internal/apperror"
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
		writeBindError(ctx, err)
		return
	}

	stock, err := ah.stockService.CreateStock(ctx, req)
	if err != nil {
		ah.logger.WithError(err).WithFields(logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"sku":        req.SKU,
		}).Error(logger.LogMessageFailedToCreateStock)
		status, code, message, details := mapDomainError(err)
		writeError(ctx, status, message, code, details)
		return
	}
	writeSuccess(ctx, http.StatusOK, "Stock created", stock)
}

// UpdateStock updates the stock quantity for a given SKU.
func (ah *AdminHandler) UpdateStock(ctx *gin.Context) {
	var req apimodel.StockRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeBindError(ctx, err)
		return
	}

	stock, err := ah.adminStockOrchestrator.AdjustInventory(ctx, req)
	if err != nil {
		ah.logger.WithError(err).WithFields(logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"sku":        req.SKU,
		}).Error(logger.LogMessageFailedToUpdateStock)
		status, code, message, details := mapDomainError(err)
		writeError(ctx, status, message, code, details)
		return
	}
	writeSuccess(ctx, http.StatusOK, "Stock updated", stock)
}

// DeleteStock deletes a stock item by SKU.
func (ah *AdminHandler) DeleteStock(ctx *gin.Context) {
	sku := ctx.Param("sku")
	if sku == "" {
		writeError(
			ctx,
			http.StatusBadRequest,
			apperror.CodeValidationErrorMessage,
			apperror.CodeValidationErrorCode,
			[]apimodel.ErrorDetail{{Field: "sku", Reason: "required"}},
		)
		return
	}
	err := ah.adminStockOrchestrator.DeleteStock(ctx, sku)
	if err != nil {
		ah.logger.WithError(err).WithFields(logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"sku":        sku,
		}).Error(logger.LogMessageFailedToDeleteStock)
		status, code, message, details := mapDomainError(err)
		writeError(ctx, status, message, code, details)
		return
	}
	writeSuccess(ctx, http.StatusOK, "Stock deleted", gin.H{"sku": sku})
}

// GetActiveReservationItemsBySku lists active reservation items by SKU with pagination.
func (ah *AdminHandler) GetActiveReservationItemsBySku(ctx *gin.Context) {
	sku := ctx.Param("sku")
	if sku == "" {
		writeError(
			ctx,
			http.StatusBadRequest,
			apperror.CodeValidationErrorMessage,
			apperror.CodeValidationErrorCode,
			[]apimodel.ErrorDetail{{Field: "sku", Reason: "required"}},
		)
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
		ah.logger.WithError(err).WithFields(logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"sku":        sku,
			"limit":      requestedLimit,
			"offset":     requestedOffset,
		}).Error(logger.LogMessageFailedToGetActiveReservation)
		status, code, msg, details := mapDomainError(err)
		writeError(ctx, status, msg, code, details)
		return
	}

	writeSuccess(ctx, http.StatusOK, message, gin.H{
		"sku":        sku,
		"items":      items,
		"pagination": pagination,
	})
}
