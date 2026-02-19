package handler

import (
	"go-inventory-reservations/internal/application"
	"go-inventory-reservations/internal/logger"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReservationHandler handles all reservation-related routes, e.g., reservation creation, etc.
type ReservationHandler struct {
	reservationOrchestrator application.ReservationOrchestratorInterface
	reservationService      service.ReservationServiceInterface
	logger                  *logger.Logger
}

// NewReservationHandler creates a new ReservationHandler instance.
func NewReservationHandler(
	reservationOrchestrator application.ReservationOrchestratorInterface,
	reservationService service.ReservationServiceInterface,
	logger *logger.Logger,
) *ReservationHandler {
	return &ReservationHandler{
		reservationOrchestrator: reservationOrchestrator,
		reservationService:      reservationService,
		logger:                  logger,
	}
}

// CreateReservation handles the creation of a new reservation based on the provided request payload.
func (rh *ReservationHandler) CreateReservation(ctx *gin.Context) {
	var req apimodel.CreateReservationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	reservation, err := rh.reservationOrchestrator.CreateReservation(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create reservation: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, reservation)
}

// UpdateReservation handles the update of an existing reservation based on the provided request payload.
func (rh *ReservationHandler) UpdateReservation(ctx *gin.Context) {
	var req apimodel.UpdateReservationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	reservation, err := rh.reservationOrchestrator.UpdateReservation(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update reservation: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, reservation)
}

// GetReservationById retrieves a reservation by its unique identifier from the request context and returns it.
func (rh *ReservationHandler) GetReservationById(ctx *gin.Context) {
	idStr := ctx.Param("id")

	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "id parameter is required",
		})
		return
	}

	ReservationId, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reservation_id format",
		})
		return
	}

	reservation, err := rh.reservationOrchestrator.GetReservationById(ctx, ReservationId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Reservation not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, reservation)
}

// GetReservationByQuoteId retrieves a reservation by its associated quote ID from the request context and returns it.
func (rh *ReservationHandler) GetReservationByQuoteId(ctx *gin.Context) {
	quoteId := ctx.Param("quote_id")

	if strings.TrimSpace(quoteId) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "quote parameter is required",
		})
		return
	}

	reservation, err := rh.reservationOrchestrator.GetReservationByQuoteId(ctx, quoteId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Reservation not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, reservation)
}

// GetReservationByOrderId retrieves a reservation by its associated order ID from the request context and returns it.
func (rh *ReservationHandler) GetReservationByOrderId(ctx *gin.Context) {
	orderId := ctx.Param("order_id")

	if strings.TrimSpace(orderId) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "order parameter is required",
		})
		return
	}

	reservation, err := rh.reservationOrchestrator.GetReservationByOrderId(ctx, orderId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Reservation not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, reservation)
}

// AttachOrder handles the request to link an existing order with a reservation based on the provided request payload.
func (rh *ReservationHandler) AttachOrder(ctx *gin.Context) {
	var req apimodel.AttachOrderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	err := rh.reservationService.AttachOrder(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to attach order to reservation: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "order attached successfully"})
}

// CommitReservation processes a request to finalize a reservation associated with a given order and marks it as committed.
func (rh *ReservationHandler) CommitReservation(ctx *gin.Context) {
	var req apimodel.CommitReservationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	_, err := rh.reservationOrchestrator.CommitReservation(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit reservation: " + err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation committed success"})
}

// ReleaseReservation handles the release of a reservation based on its ID provided in the request context.
func (rh *ReservationHandler) ReleaseReservation(ctx *gin.Context) {
	idStr := ctx.Param("id")

	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "id parameter is required",
		})
		return
	}

	ReservationId, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reservation_id format",
		})
		return
	}

	err = rh.reservationOrchestrator.ReleaseReservation(ctx, ReservationId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to release reservation: " + err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Reservation released"})
}

// Revert handles the request to revert a reservation based on the provided request payload.
func (rh *ReservationHandler) Revert(ctx *gin.Context) {
	var req apimodel.RevertReservationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	err := rh.reservationOrchestrator.RevertReservation(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to revert reservation: " + err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation revert success"})
}
