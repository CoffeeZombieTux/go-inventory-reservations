package application

import (
	"context"
	"go-inventory-reservations/internal/apperror"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"
)

// CreateReservation orchestrates the process of creating a reservation.
// It creates the reservation and reservation items.
// It also updates the stock inventory.
func (ro *ReservationOrchestrator) CreateReservation(
	ctx context.Context,
	params apimodel.CreateReservationRequest,
) (*apimodel.ReservationResponse, error) {
	result := apimodel.ReservationResponse{Items: make(map[string]*model.ReservationItem)}

	items := params.Items

	err := uow.WithUnitOfWork(ctx, ro.uowManager, func(unit *uow.UnitOfWork) error {
		// Check if there are enough stock items for all items in the reservation
		for _, item := range items {
			if item.Quantity == 0 {
				return apperror.New(
					apperror.CodeValidationError,
					"Quantity must be greater than 0",
					"qty must be greater than 0",
				)
			}
			stock, err := ro.stockService.GetStockBySkuForUpdate(ctx, item.SKU, unit)
			if err != nil {
				return err
			}
			if ro.stockService.CalculateAvailability(ctx, stock) < item.Quantity {
				return apperror.New(
					apperror.CodeInsufficientStock,
					"Insufficient stock for requested items",
					"insufficient stock for SKU "+item.SKU,
				)
			}
		}

		// Create reservation
		reservation, err := ro.reservationService.CreateReservationHelper(ctx, params, unit)
		if err != nil {
			return err
		}
		result.Reservation = reservation

		for _, item := range items {
			// Create reservation items
			createdItem, err := ro.reservationItemService.CreateReservationItem(
				ctx,
				item,
				reservation.ReservationId,
				unit,
			)
			if err != nil {
				return err
			}

			result.Items[item.SKU] = createdItem

			// Update stock reserved quantity
			_, err = ro.stockService.ReserveStock(ctx, item.SKU, item.Quantity, unit)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}
