package handler

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/service"
	"net/http"
)

// ReservationHandler handles all reservation-related routes, e.g. reservation creation, etc.
type ReservationHandler struct {
	reservationService service.ReservationServiceInterface
	logger             *logger.Logger
}

// NewReservationHandler creates a new ReservationHandler instance.
func NewReservationHandler(reservationService service.ReservationServiceInterface, logger *logger.Logger) *ReservationHandler {
	return &ReservationHandler{
		reservationService: reservationService,
		logger:             logger,
	}
}

// TO be completed
func (rh *ReservationHandler) CreateReservation(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation created"})
}

// TO be completed
func (rh *ReservationHandler) UpdateReservation(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation updated"})
}

// TO be completed
func (rh *ReservationHandler) GetReservation(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "get reservation success"})
}

// TO be completed
func (rh *ReservationHandler) GetReservationByQuoteId(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "get reservation by quote ID success"})
}

// TO be completed
func (rh *ReservationHandler) GetReservationByOrderId(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "get reservation by order ID success"})
}

// TO be completed
func (rh *ReservationHandler) DeleteReservation(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation deleted"})
}

// TO be completed
func (rh *ReservationHandler) CommitReservation(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation committed success"})
}

// TO be completed
func (rh *ReservationHandler) Attach(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation attached success"})
}

// TO be completed
func (rh *ReservationHandler) GetReservationAvailability(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation availability success"})
}
