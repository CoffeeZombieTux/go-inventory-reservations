package service

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/repository"
	"go-inventory-reservations/internal/uow"
	"time"

	"github.com/google/uuid"
)

// initial statuses
const statusPending = "PENDING"
const statusExpired = "EXPIRED"
const statusReserved = "RESERVED"

// final statuses
const statusCommitted = "COMMITTED"
const statusReleased = "RELEASED"
const statusReverted = "REVERTED"

// ReservationService defines business operations for reservation management
type ReservationServiceInterface interface {
	CreateReservation(ctx context.Context, request apimodel.CreateReservationRequest, uow *uow.UnitOfWork) (*model.Reservation, error)
	UpdateReservation(ctx context.Context, request apimodel.UpdateReservationRequest, uow *uow.UnitOfWork) (*model.Reservation, error)
	GetReservationById(ctx context.Context, id uuid.UUID) (*model.Reservation, error)
	GetReservationByQuoteId(ctx context.Context, quoteId string) (*model.Reservation, error)
	GetReservationByOrderId(ctx context.Context, orderId string) (*model.Reservation, error)
	GetToBeExpiredReservations(ctx context.Context) ([]*model.Reservation, error)
	AttachOrder(ctx context.Context, request apimodel.AttachOrderRequest) error
	CommitReservation(ctx context.Context, reservation *model.Reservation, orderId string, uow *uow.UnitOfWork) (*model.Reservation, error)
	ReleaseReservation(ctx context.Context, id uuid.UUID, uow *uow.UnitOfWork) error
	RevertReservation(ctx context.Context, request apimodel.RevertReservationRequest, uow *uow.UnitOfWork) error
	ExpireReservation(ctx context.Context, reservation *model.Reservation, uow *uow.UnitOfWork) error
	ArchiveReservations(ctx context.Context) (int, error)
	getExpiresAt() *time.Time
}

// ReservationService is a service for reservation management.
type ReservationService struct {
	repo   repository.ReservationRepositoryInterface
	config *config.Config
}

// NewReservationService creates a new ReservationService instance.
func NewReservationService(
	repo repository.ReservationRepositoryInterface,
	config *config.Config,
) ReservationServiceInterface {
	return &ReservationService{
		repo:   repo,
		config: config,
	}
}

// CreateReservation creates a new reservation.
func (rs ReservationService) CreateReservation(
	ctx context.Context,
	request apimodel.CreateReservationRequest,
	uow *uow.UnitOfWork,
) (*model.Reservation, error) {

	reservation := &model.Reservation{
		ReservationId: uuid.New(),
		Status:        statusPending,
		QuoteId:       &request.QuoteId,
		ExpiresAt:     rs.getExpiresAt(),
	}

	result, err := rs.repo.Save(ctx, reservation, uow)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateReservation updates an existing reservation.
func (rs ReservationService) UpdateReservation(
	ctx context.Context,
	request apimodel.UpdateReservationRequest,
	uow *uow.UnitOfWork,
) (*model.Reservation, error) {
	reservation, err := rs.repo.GetById(ctx, *request.ReservationId)
	if err != nil {
		return nil, err
	}

	err = checkAvailableStatuses(*reservation, []string{statusPending, statusExpired})
	if err != nil {
		return nil, err
	}
	reservation.QuoteId = &request.QuoteId
	reservation.Status = statusPending
	reservation.ExpiresAt = rs.getExpiresAt()

	result, err := rs.repo.Save(ctx, reservation, uow)

	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetReservationById returns a reservation by its ID.
func (rs ReservationService) GetReservationById(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	reservation, err := rs.repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// GetReservationByQuoteId returns a reservation by its quote ID.
func (rs ReservationService) GetReservationByQuoteId(ctx context.Context, quoteId string) (*model.Reservation, error) {
	reservation, err := rs.repo.GetByQuoteId(ctx, quoteId)
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// GetReservationByOrderId returns a reservation by its order ID.
func (rs ReservationService) GetReservationByOrderId(ctx context.Context, orderId string) (*model.Reservation, error) {
	reservation, err := rs.repo.GetByOrderId(ctx, orderId)
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// AttachOrder attaches an order to a reservation.
func (rs ReservationService) AttachOrder(ctx context.Context, request apimodel.AttachOrderRequest) error {
	reservation, err := rs.repo.GetById(ctx, *request.ReservationId)
	if err != nil {
		return err
	}

	err = checkAvailableStatuses(*reservation, []string{statusPending})

	reservation.OrderId = &request.OrderId
	reservation.Status = statusReserved
	reservation.ExpiresAt = nil

	_, err = rs.repo.Save(ctx, reservation, nil)
	if err != nil {
		return err
	}
	return nil
}

// CommitReservation marks a reservation as committed.
func (rs ReservationService) CommitReservation(
	ctx context.Context,
	reservation *model.Reservation,
	orderId string,
	uow *uow.UnitOfWork,
) (*model.Reservation, error) {

	if *reservation.OrderId != orderId {
		err := fmt.Errorf(
			"reservation order id %s does not match request order id %s",
			reservation.OrderId,
			orderId,
		)
		return nil, err
	}

	err := checkAvailableStatuses(*reservation, []string{statusReserved})

	reservation.OrderId = &orderId
	reservation.Status = statusCommitted
	reservation.ExpiresAt = nil

	reservation, err = rs.repo.Save(ctx, reservation, uow)
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// ReleaseReservation marks a reservation as released.
func (rs ReservationService) ReleaseReservation(ctx context.Context, id uuid.UUID, uow *uow.UnitOfWork) error {
	reservation, err := rs.repo.GetById(ctx, id)
	if err != nil {
		return err
	}
	err = checkAvailableStatuses(*reservation, []string{statusReserved, statusPending, statusExpired})
	if err != nil {
		return err
	}
	reservation.Status = statusReleased
	reservation.ExpiresAt = nil

	_, err = rs.repo.Save(ctx, reservation, uow)
	if err != nil {
		return err
	}
	return nil
}

// RevertReservation marks a reservation as reverted.
func (rs ReservationService) RevertReservation(
	ctx context.Context,
	request apimodel.RevertReservationRequest,
	uow *uow.UnitOfWork,
) error {
	reservation, err := rs.repo.GetById(ctx, *request.ReservationId)
	if err != nil {
		return err
	}
	if *reservation.OrderId != request.OrderId {
		err = fmt.Errorf(
			"reservation order id %s does not match request order id %s",
			reservation.OrderId,
			request.OrderId,
		)
		return err
	}

	err = checkAvailableStatuses(*reservation, []string{statusCommitted})

	reservation.Status = statusReverted

	_, err = rs.repo.Save(ctx, reservation, uow)
	if err != nil {
		return err
	}
	return nil
}

func (rs ReservationService) GetToBeExpiredReservations(ctx context.Context) ([]*model.Reservation, error) {
	now := time.Now()
	query := apimodel.ReservationsQuery{
		ExpiresAtGte: &now,
		Statuses:     []string{statusPending},
	}
	reservations, err := rs.repo.SelectReservationsByQuery(ctx, query, rs.config.QuoteExpirationSettings.Limit)
	if err != nil {
		return nil, err
	}
	return reservations, nil
	// later there will be logic with updating reserved qty ane delete items

}

func (rs ReservationService) ExpireReservation(
	ctx context.Context,
	reservation *model.Reservation,
	uow *uow.UnitOfWork,
) error {
	reservation.Status = statusExpired
	_, err := rs.repo.Save(ctx, reservation, uow)
	return err
}

// ArchiveReservations deletes reservations older than ArchiveSettings.ArchiveReservationsAfterDays days.
func (rs ReservationService) ArchiveReservations(ctx context.Context) (int, error) {
	archivaAfter := time.Now().AddDate(0, 0, -rs.config.ArchiveSettings.ArchiveReservationsAfterDays)

	query := apimodel.ReservationsQuery{
		Statuses:    []string{statusCommitted, statusReleased, statusReverted, statusExpired},
		UpdatedAtLt: &archivaAfter,
	}

	reservations, err := rs.repo.SelectReservationsByQuery(ctx, query, rs.config.ArchiveSettings.Limit)
	if err != nil {
		return 0, err
	}
	// later there will be logic with moving to the archive table or other archive strategy
	for _, reservation := range reservations {
		err = rs.repo.Delete(ctx, reservation.ReservationId)
		if err != nil {
			return 0, err
		}
	}
	return len(reservations), nil
}

// getExpiresAt returns reservation expiration time based on QuoteExpirationSettings.QuoteExpirationMinutes.
func (rs ReservationService) getExpiresAt() *time.Time {
	res := time.Now().Add(time.Duration(rs.config.QuoteExpirationSettings.QuoteExpirationMinutes) * time.Minute)
	return &res
}

// checkAvailableStatuses checks if reservation status is allowed for the requested action.
func checkAvailableStatuses(reservation model.Reservation, allowedStatuses []string) error {
	for _, status := range allowedStatuses {
		if reservation.Status == status {
			return nil
		}
	}
	return fmt.Errorf("reservation status %s is not allowed for this action", reservation.Status)
}
