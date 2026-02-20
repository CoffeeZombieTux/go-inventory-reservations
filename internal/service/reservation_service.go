package service

import (
	"context"
	"errors"
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

// ReservationServiceInterface ReservationService defines business operations for reservation management
// Helper methods are used to avoid exposing repository methods directly to the controller.
// Helpers do not do all necessary changes on other entities
type ReservationServiceInterface interface {
	GetReservationById(ctx context.Context, id uuid.UUID) (*model.Reservation, error)
	GetReservationByIdForUpdate(ctx context.Context, id uuid.UUID, uow *uow.UnitOfWork) (*model.Reservation, error)
	GetReservationByQuoteId(ctx context.Context, quoteId string) (*model.Reservation, error)
	GetReservationByOrderId(ctx context.Context, orderId string) (*model.Reservation, error)
	GetToBeExpiredReservations(ctx context.Context) ([]*model.Reservation, error)
	AttachOrder(ctx context.Context, request apimodel.AttachOrderRequest) error
	ArchiveReservations(ctx context.Context) (int, error)
	CreateReservationHelper(ctx context.Context, request apimodel.CreateReservationRequest, uow *uow.UnitOfWork) (*model.Reservation, error)
	UpdateReservationHelper(ctx context.Context, request apimodel.UpdateReservationRequest, uow *uow.UnitOfWork) (*model.Reservation, error)
	CommitReservationHelper(ctx context.Context, request apimodel.CommitReservationRequest, uow *uow.UnitOfWork) (*model.Reservation, error)
	ReleaseReservationHelper(ctx context.Context, id uuid.UUID, uow *uow.UnitOfWork) (*model.Reservation, error)
	RevertReservationHelper(ctx context.Context, request apimodel.RevertReservationRequest, uow *uow.UnitOfWork) (*model.Reservation, error)
	ExpireReservationHelper(ctx context.Context, reservation *model.Reservation, uow *uow.UnitOfWork) error
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

// GetReservationById returns a reservation by its ID.
func (rs ReservationService) GetReservationById(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	reservation, err := rs.repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// GetReservationByIdForUpdate returns a reservation by its ID with FOR UPDATE lock.
func (rs ReservationService) GetReservationByIdForUpdate(ctx context.Context, id uuid.UUID, uow *uow.UnitOfWork) (*model.Reservation, error) {
	reservation, err := rs.repo.GetByIdForUpdate(ctx, id, uow)
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

// GetToBeExpiredReservations returns pending reservations eligible for expiration processing.
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

// CreateReservationHelper creates a new reservation record.
func (rs ReservationService) CreateReservationHelper(
	ctx context.Context,
	request apimodel.CreateReservationRequest,
	uow *uow.UnitOfWork,
) (*model.Reservation, error) {

	reservation := &model.Reservation{
		ReservationId: uuid.New(),
		Status:        statusPending,
		QuoteId:       request.QuoteId,
		ExpiresAt:     rs.getExpiresAt(),
		ItemsHash:     BuildReservationItemsHashFromRequests(request.Items),
	}

	result, err := rs.repo.Save(ctx, reservation, uow)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateReservationHelper updates an existing reservation.
func (rs ReservationService) UpdateReservationHelper(
	ctx context.Context,
	request apimodel.UpdateReservationRequest,
	uow *uow.UnitOfWork,
) (*model.Reservation, error) {
	reservation, err := rs.repo.GetByIdForUpdate(ctx, *request.ReservationId, uow)
	if err != nil {
		return nil, err
	}
	err = checkAvailableStatuses(*reservation, []string{statusPending, statusExpired})
	if err != nil {
		return nil, err
	}

	requestItemsHash := BuildReservationItemsHashFromRequests(request.Items)
	if reservation.QuoteId == request.QuoteId &&
		itemsHashEqual(reservation.ItemsHash, requestItemsHash) {
		// Nothing to update
		return reservation, nil
	}

	reservation.QuoteId = request.QuoteId
	reservation.Status = statusPending
	reservation.ExpiresAt = rs.getExpiresAt()
	reservation.ItemsHash = requestItemsHash

	result, err := rs.repo.Save(ctx, reservation, uow)

	if err != nil {
		return nil, err
	}
	return result, nil
}

// CommitReservationHelper marks a reservation as committed.
func (rs ReservationService) CommitReservationHelper(
	ctx context.Context,
	request apimodel.CommitReservationRequest,
	uow *uow.UnitOfWork,
) (*model.Reservation, error) {
	reservation, err := rs.repo.GetByIdForUpdate(ctx, *request.ReservationId, uow)
	if err != nil {
		return nil, err
	}

	if reservation.OrderId == nil {
		return nil, fmt.Errorf("reservation %s has no attached order", reservation.ReservationId)
	}

	if *reservation.OrderId != request.OrderId {
		err := fmt.Errorf(
			"reservation order id %s does not match request order id %s",
			*reservation.OrderId,
			request.OrderId,
		)
		return nil, err
	}

	err = checkAvailableStatuses(*reservation, []string{statusReserved})
	if err != nil {
		return nil, err
	}

	reservation.OrderId = &request.OrderId
	reservation.Status = statusCommitted
	reservation.ExpiresAt = nil
	reservation.ItemsHash = nil

	reservation, err = rs.repo.Save(ctx, reservation, uow)
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// ReleaseReservationHelper marks a reservation as released.
func (rs ReservationService) ReleaseReservationHelper(
	ctx context.Context,
	id uuid.UUID,
	uow *uow.UnitOfWork,
) (*model.Reservation, error) {
	reservation, err := rs.repo.GetByIdForUpdate(ctx, id, uow)
	if err != nil {
		return nil, err
	}
	err = checkAvailableStatuses(*reservation, []string{statusReserved, statusPending, statusExpired})
	if err != nil {
		return nil, err
	}
	reservation.Status = statusReleased
	reservation.ExpiresAt = nil
	reservation.ItemsHash = nil

	_, err = rs.repo.Save(ctx, reservation, uow)
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// RevertReservationHelper marks a reservation as reverted.
func (rs ReservationService) RevertReservationHelper(
	ctx context.Context,
	request apimodel.RevertReservationRequest,
	uow *uow.UnitOfWork,
) (*model.Reservation, error) {
	reservation, err := rs.repo.GetByIdForUpdate(ctx, *request.ReservationId, uow)
	if err != nil {
		return nil, err
	}
	if *reservation.OrderId != request.OrderId {
		err = fmt.Errorf(
			"reservation order id %s does not match request order id %s",
			*reservation.OrderId,
			request.OrderId,
		)
		return nil, err
	}

	err = checkAvailableStatuses(*reservation, []string{statusCommitted})

	reservation.Status = statusReverted
	reservation.ItemsHash = nil

	_, err = rs.repo.Save(ctx, reservation, uow)
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// ExpireReservationHelper marks a reservation as expired and clears its items hash.
func (rs ReservationService) ExpireReservationHelper(
	ctx context.Context,
	reservation *model.Reservation,
	uow *uow.UnitOfWork,
) error {
	reservation.Status = statusExpired
	reservation.ItemsHash = nil
	_, err := rs.repo.Save(ctx, reservation, uow)
	return err
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

// IsReservationVersionConflict reports whether the error is an optimistic lock conflict.
func IsReservationVersionConflict(err error) bool {
	return errors.Is(err, repository.ErrReservationVersionConflict)
}

// itemsHashEqual compares two optional item hashes.
func itemsHashEqual(a *string, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
