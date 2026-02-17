package application

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
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
	// Check if there are enough stock items for all items in the reservation
	for _, item := range items {
		stock, err := ro.stockService.GetStockBySku(ctx, item.SKU)
		if err != nil {
			return nil, err
		}
		if ro.stockService.CalculateAvailability(ctx, stock) < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for SKU %s", item.SKU)
		}
	}

	unit, err := ro.uowManager.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			unit.Rollback()
		}
	}()

	// Create reservation
	reservation, err := ro.reservationService.CreateReservation(ctx, params, unit)
	result.Reservation = reservation
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		// Create reservation items
		createdItem, err := ro.reservationItemService.CreateReservationItem(
			ctx,
			item,
			reservation.ReservationId,
			unit,
		)
		if err != nil {
			return nil, err
		}

		result.Items[item.SKU] = createdItem

		// Update stock reserved quantity
		_, err = ro.stockService.ReserveStock(ctx, item.SKU, item.Quantity, unit)
		if err != nil {
			return nil, err
		}
	}

	err = unit.Commit()
	if err != nil {
		return nil, err
	}
	return &result, nil
}
