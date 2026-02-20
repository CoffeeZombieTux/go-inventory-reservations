package service

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/repository"
	"go-inventory-reservations/internal/uow"

	"github.com/google/uuid"
)

// ReservationItemsServiceInterface defines operations for reservation items management
type ReservationItemsServiceInterface interface {
	GetReservationItem(ctx context.Context, reservationId uuid.UUID, sku string, uow *uow.UnitOfWork) (*model.ReservationItem, error)
	GetReservationItems(ctx context.Context, reservationId uuid.UUID, uow *uow.UnitOfWork) (map[string]*model.ReservationItem, error)
	CreateReservationItem(ctx context.Context, request apimodel.ReservationItemRequest, reservationId uuid.UUID, uow *uow.UnitOfWork) (*model.ReservationItem, error)
	UpdateReservationItem(ctx context.Context, request apimodel.ReservationItemRequest, reservationId uuid.UUID, uow *uow.UnitOfWork) (*model.ReservationItem, error)
	SetReservationItemActive(ctx context.Context, reservationId uuid.UUID, sku string, isActive bool, uow *uow.UnitOfWork) (*model.ReservationItem, error)
	DeleteReservationItem(ctx context.Context, reservationId uuid.UUID, sku string, uow *uow.UnitOfWork) error
}

// ReservationItemsService is a service for reservation items management.
type ReservationItemsService struct {
	repo   repository.ReservationItemsRepositoryInterface
	config *config.Config
}

// NewReservationItemsService creates a new ReservationItemsService instance.
func NewReservationItemsService(
	repo repository.ReservationItemsRepositoryInterface,
	config *config.Config,
) ReservationItemsServiceInterface {
	return &ReservationItemsService{
		repo:   repo,
		config: config,
	}
}

// GetReservationItem retrieves a reservation item by reservation ID and SKU using the provided context and unit of work if available.
func (ris ReservationItemsService) GetReservationItem(
	ctx context.Context,
	reservationId uuid.UUID,
	sku string,
	uow *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	reservationItem, err := ris.repo.Get(ctx, reservationId, sku, uow)
	if err != nil {
		return nil, err
	}
	return reservationItem, nil
}

// GetReservationItems retrieves all reservation items for a given reservation ID using the provided context and unit of work if available.
func (ris ReservationItemsService) GetReservationItems(
	ctx context.Context,
	reservationId uuid.UUID,
	uow *uow.UnitOfWork,
) (map[string]*model.ReservationItem, error) {
	reservationItems, err := ris.repo.FindByReservationId(ctx, reservationId, uow)
	if err != nil {
		return nil, err
	}
	return reservationItems, nil
}

// CreateReservationItem creates a new reservation item based on the given request and associates it with a reservation ID.
func (ris ReservationItemsService) CreateReservationItem(
	ctx context.Context,
	request apimodel.ReservationItemRequest,
	reservationId uuid.UUID,
	uow *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	reservationItem := &model.ReservationItem{
		ReservationId: reservationId,
		SKU:           request.SKU,
		Qty:           request.Quantity,
		IsActive:      true,
	}

	err := ris.validateReservationItemQuantity(reservationItem)
	if err != nil {
		return nil, err
	}

	result, err := ris.repo.Create(ctx, reservationItem, uow)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateReservationItem updates an existing reservation item based on the given request and associates it with a reservation ID.
func (ris ReservationItemsService) UpdateReservationItem(
	ctx context.Context,
	request apimodel.ReservationItemRequest,
	reservationId uuid.UUID,
	uow *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	reservationItem, err := ris.repo.Get(ctx, reservationId, request.SKU, uow)
	if err != nil {
		return nil, err
	}
	reservationItem.Qty = request.Quantity
	reservationItem.IsActive = request.Quantity > 0

	err = ris.validateReservationItemQuantity(reservationItem)
	if err != nil {
		return nil, err
	}

	result, err := ris.repo.Update(ctx, reservationItem, uow)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SetReservationItemActive updates reservation item active flag.
func (ris ReservationItemsService) SetReservationItemActive(
	ctx context.Context,
	reservationId uuid.UUID,
	sku string,
	isActive bool,
	uow *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	reservationItem, err := ris.repo.SetIsActive(ctx, reservationId, sku, isActive, uow)
	if err != nil {
		return nil, err
	}
	return reservationItem, nil
}

// DeleteReservationItem deletes a reservation item based on the given reservation ID and SKU.
func (ris ReservationItemsService) DeleteReservationItem(
	ctx context.Context,
	reservationId uuid.UUID,
	sku string,
	uow *uow.UnitOfWork,
) error {
	err := ris.repo.Delete(ctx, reservationId, sku, uow)
	if err != nil {
		return err
	}
	return nil
}

// validateReservationItemQuantity validates reservation item quantity constraints.
func (ris ReservationItemsService) validateReservationItemQuantity(reservationItem *model.ReservationItem) error {
	if reservationItem.Qty > ris.config.ReservationItemSettings.MaxQuantity {
		return fmt.Errorf("quantity exceeds maximum allowed: %d", ris.config.ReservationItemSettings.MaxQuantity)
	}
	if reservationItem.Qty < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}
	return nil
}
