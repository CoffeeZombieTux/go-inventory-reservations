package handler

import (
	"go-inventory-reservations/internal/application"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/service"
)

// HandlersPool represents a collection of handlers
type HandlersPool struct {
	Stock       *StockHandler
	Reservation *ReservationHandler
	Admin       *AdminHandler
}

// NewHandlersPool creates a new HandlersPool instance
func NewHandlersPool(
	reservationOrchestrator application.ReservationOrchestratorInterface,
	stockService service.StockServiceInterface,
	reservationService service.ReservationServiceInterface,
	logger *logger.Logger,
) *HandlersPool {
	return &HandlersPool{
		Stock:       NewStockHandler(stockService, logger),
		Reservation: NewReservationHandler(reservationOrchestrator, reservationService, logger),
		Admin:       NewAdminHandler(stockService, logger),
	}
}
