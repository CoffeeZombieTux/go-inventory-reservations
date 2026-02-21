package handler

import (
	"go-inventory-reservations/internal/apperror"
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
		writeBindError(ctx, err)
		return
	}

	reservation, err := rh.reservationOrchestrator.CreateReservation(ctx, req)
	if err != nil {
		rh.logger.WithError(err).WithFields(logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"quote_id":   req.QuoteId,
		}).Error(logger.LogMessageFailedToCreateReservation)
		status, code, message, details := mapDomainError(err)
		writeError(ctx, status, message, code, details)
		return
	}

	writeSuccess(ctx, http.StatusOK, "Reservation created", reservation)
}

// UpdateReservation handles the update of an existing reservation based on the provided request payload.
func (rh *ReservationHandler) UpdateReservation(ctx *gin.Context) {
	var req apimodel.UpdateReservationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeBindError(ctx, err)
		return
	}

	reservation, err := rh.reservationOrchestrator.UpdateReservation(ctx, req)
	if err != nil {
		fields := logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"quote_id":   req.QuoteId,
		}
		if req.ReservationId != nil {
			fields["reservation_id"] = req.ReservationId.String()
		}
		rh.logger.WithError(err).WithFields(fields).Error(logger.LogMessageFailedToUpdateReservation)
		status, code, message, details := mapDomainError(err)
		writeError(ctx, status, message, code, details)
		return
	}

	writeSuccess(ctx, http.StatusOK, "Reservation updated", reservation)
}

// GetReservationById retrieves a reservation by its unique identifier from the request context and returns it.
func (rh *ReservationHandler) GetReservationById(ctx *gin.Context) {
	idStr := ctx.Param("id")

	if idStr == "" {
		writeError(
			ctx,
			http.StatusBadRequest,
			apperror.CodeValidationErrorMessage,
			apperror.CodeValidationErrorCode,
			[]apimodel.ErrorDetail{{Field: "id", Reason: "required"}},
		)
		return
	}

	ReservationId, err := uuid.Parse(idStr)
	if err != nil {
		writeError(
			ctx,
			http.StatusBadRequest,
			apperror.CodeValidationErrorMessage,
			apperror.CodeValidationErrorCode,
			[]apimodel.ErrorDetail{{Field: "id", Reason: "uuid"}},
		)
		return
	}

	reservation, err := rh.reservationOrchestrator.GetReservationById(ctx, ReservationId)
	if err != nil {
		rh.logger.WithError(err).WithFields(logger.Fields{
			"request_id":     requestIDFromContext(ctx),
			"reservation_id": ReservationId.String(),
		}).Error(logger.LogMessageFailedToGetReservationByID)
		writeError(ctx, http.StatusNotFound, apperror.CodeNotFoundMessage, apperror.CodeNotFoundCode, nil)
		return
	}
	writeSuccess(ctx, http.StatusOK, "Reservation fetched", reservation)
}

// GetReservationByQuoteId retrieves a reservation by its associated quote ID from the request context and returns it.
func (rh *ReservationHandler) GetReservationByQuoteId(ctx *gin.Context) {
	quoteId := ctx.Param("quote_id")

	if strings.TrimSpace(quoteId) == "" {
		writeError(
			ctx,
			http.StatusBadRequest,
			apperror.CodeValidationErrorMessage,
			apperror.CodeValidationErrorCode,
			[]apimodel.ErrorDetail{{Field: "quote_id", Reason: "required"}},
		)
		return
	}

	reservation, err := rh.reservationOrchestrator.GetReservationByQuoteId(ctx, quoteId)
	if err != nil {
		rh.logger.WithError(err).WithFields(logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"quote_id":   quoteId,
		}).Error(logger.LogMessageFailedToGetReservationByQuoteID)
		writeError(ctx, http.StatusNotFound, apperror.CodeNotFoundMessage, apperror.CodeNotFoundCode, nil)
		return
	}
	writeSuccess(ctx, http.StatusOK, "Reservation fetched", reservation)
}

// GetReservationByOrderId retrieves a reservation by its associated order ID from the request context and returns it.
func (rh *ReservationHandler) GetReservationByOrderId(ctx *gin.Context) {
	orderId := ctx.Param("order_id")

	if strings.TrimSpace(orderId) == "" {
		writeError(
			ctx,
			http.StatusBadRequest,
			apperror.CodeValidationErrorMessage,
			apperror.CodeValidationErrorCode,
			[]apimodel.ErrorDetail{{Field: "order_id", Reason: "required"}},
		)
		return
	}

	reservation, err := rh.reservationOrchestrator.GetReservationByOrderId(ctx, orderId)
	if err != nil {
		rh.logger.WithError(err).WithFields(logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"order_id":   orderId,
		}).Error(logger.LogMessageFailedToGetReservationByOrderID)
		writeError(ctx, http.StatusNotFound, apperror.CodeNotFoundMessage, apperror.CodeNotFoundCode, nil)
		return
	}
	writeSuccess(ctx, http.StatusOK, "Reservation fetched", reservation)
}

// AttachOrder handles the request to link an existing order with a reservation based on the provided request payload.
func (rh *ReservationHandler) AttachOrder(ctx *gin.Context) {
	var req apimodel.AttachOrderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeBindError(ctx, err)
		return
	}

	err := rh.reservationService.AttachOrder(ctx, req)
	if err != nil {
		fields := logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"order_id":   req.OrderId,
		}
		if req.ReservationId != nil {
			fields["reservation_id"] = req.ReservationId.String()
		}
		rh.logger.WithError(err).WithFields(fields).Error(logger.LogMessageFailedToAttachOrder)
		status, code, message, details := mapDomainError(err)
		writeError(ctx, status, message, code, details)
		return
	}

	writeSuccess(ctx, http.StatusOK, "Order attached successfully", nil)
}

// CommitReservation processes a request to finalize a reservation associated with a given order and marks it as committed.
func (rh *ReservationHandler) CommitReservation(ctx *gin.Context) {
	var req apimodel.CommitReservationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeBindError(ctx, err)
		return
	}

	_, err := rh.reservationOrchestrator.CommitReservation(ctx, req)
	if err != nil {
		fields := logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"order_id":   req.OrderId,
		}
		if req.ReservationId != nil {
			fields["reservation_id"] = req.ReservationId.String()
		}
		rh.logger.WithError(err).WithFields(fields).Error(logger.LogMessageFailedToCommitReservation)
		status, code, message, details := mapDomainError(err)
		writeError(ctx, status, message, code, details)
		return
	}
	writeSuccess(ctx, http.StatusOK, "Reservation committed", nil)
}

// ReleaseReservation handles the release of a reservation based on its ID provided in the request context.
func (rh *ReservationHandler) ReleaseReservation(ctx *gin.Context) {
	idStr := ctx.Param("id")

	if idStr == "" {
		writeError(
			ctx,
			http.StatusBadRequest,
			apperror.CodeValidationErrorMessage,
			apperror.CodeValidationErrorCode,
			[]apimodel.ErrorDetail{{Field: "id", Reason: "required"}},
		)
		return
	}

	ReservationId, err := uuid.Parse(idStr)
	if err != nil {
		writeError(
			ctx,
			http.StatusBadRequest,
			apperror.CodeValidationErrorMessage,
			apperror.CodeValidationErrorCode,
			[]apimodel.ErrorDetail{{Field: "id", Reason: "uuid"}},
		)
		return
	}

	err = rh.reservationOrchestrator.ReleaseReservation(ctx, ReservationId)
	if err != nil {
		rh.logger.WithError(err).WithFields(logger.Fields{
			"request_id":     requestIDFromContext(ctx),
			"reservation_id": ReservationId.String(),
		}).Error(logger.LogMessageFailedToReleaseReservation)
		status, code, message, details := mapDomainError(err)
		writeError(ctx, status, message, code, details)
		return
	}
	writeSuccess(ctx, http.StatusOK, "Reservation released", nil)
}

// Revert handles the request to revert a reservation based on the provided request payload.
func (rh *ReservationHandler) Revert(ctx *gin.Context) {
	var req apimodel.RevertReservationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeBindError(ctx, err)
		return
	}

	err := rh.reservationOrchestrator.RevertReservation(ctx, req)
	if err != nil {
		fields := logger.Fields{
			"request_id": requestIDFromContext(ctx),
			"order_id":   req.OrderId,
		}
		if req.ReservationId != nil {
			fields["reservation_id"] = req.ReservationId.String()
		}
		rh.logger.WithError(err).WithFields(fields).Error(logger.LogMessageFailedToRevertReservation)
		status, code, message, details := mapDomainError(err)
		writeError(ctx, status, message, code, details)
		return
	}
	writeSuccess(ctx, http.StatusOK, "Reservation reverted", nil)
}
