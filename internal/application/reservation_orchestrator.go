package application

import (
	"context"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/service"
	"go-inventory-reservations/internal/uow"

	"github.com/google/uuid"
)

// ReservationOrchestratorInterface represents the orchestration layer for reservation operations.
type ReservationOrchestratorInterface interface {
	CreateReservation(ctx context.Context, params apimodel.CreateReservationRequest) (*apimodel.ReservationResponse, error)
	UpdateReservation(ctx context.Context, params apimodel.UpdateReservationRequest) (*apimodel.ReservationResponse, error)
	GetReservationById(ctx context.Context, reservationId uuid.UUID) (*apimodel.ReservationResponse, error)
	GetReservationByQuoteId(ctx context.Context, quoteId string) (*apimodel.ReservationResponse, error)
	GetReservationByOrderId(ctx context.Context, orderId string) (*apimodel.ReservationResponse, error)
	CommitReservation(ctx context.Context, params apimodel.CommitReservationRequest) (*model.Reservation, error)
	ReleaseReservation(ctx context.Context, id uuid.UUID) error
	RevertReservation(ctx context.Context, request apimodel.RevertReservationRequest) error
	ProcessExpiredReservations(ctx context.Context) (successCounter int, failureCounter int, err error)
}

// ReservationOrchestrator represents the orchestration layer for reservation operations.
type ReservationOrchestrator struct {
	uowManager             *uow.UnitOfWorkManager
	stockService           service.StockServiceInterface
	reservationService     service.ReservationServiceInterface
	reservationItemService service.ReservationItemsServiceInterface
}

// NewReservationOrchestrator creates a new ReservationOrchestrator instance.
func NewReservationOrchestrator(
	uowManager *uow.UnitOfWorkManager,
	stockService service.StockServiceInterface,
	reservationService service.ReservationServiceInterface,
	reservationItemService service.ReservationItemsServiceInterface,
) ReservationOrchestratorInterface {
	return &ReservationOrchestrator{
		uowManager:             uowManager,
		stockService:           stockService,
		reservationService:     reservationService,
		reservationItemService: reservationItemService,
	}
}

// withUnitOfWork executes fn in a transaction and commits on success.
func (ro *ReservationOrchestrator) withUnitOfWork(
	ctx context.Context,
	fn func(unit *uow.UnitOfWork) error,
) (err error) {
	unit, err := ro.uowManager.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if panicErr := recover(); panicErr != nil {
			_ = unit.Rollback()
			panic(panicErr)
		}
		if err != nil {
			_ = unit.Rollback()
		}
	}()

	err = fn(unit)
	if err != nil {
		return err
	}

	return unit.Commit()
}
