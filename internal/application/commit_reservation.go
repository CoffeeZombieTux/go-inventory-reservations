package application

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
)

// CommitReservation orchestrates the process of committing a reservation.
// It updates the stock inventory and updates the reservation status.
func (ro *ReservationOrchestrator) CommitReservation(
	ctx context.Context,
	params apimodel.CommitReservationRequest,
) (*model.Reservation, error) {
	reservation, err := ro.reservationService.GetReservationById(ctx, *params.ReservationId)
	if err != nil {
		return nil, err
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

	// Get reservation items
	reservedItems, err := ro.reservationItemService.GetReservationItems(ctx, reservation.ReservationId, unit)
	if err != nil {
		return nil, fmt.Errorf(
			"you cannot commit reservation %s because it has not been reserved items",
			reservation.ReservationId.String(),
		)
	}

	for _, item := range reservedItems {
		// Get stock
		stock, err := ro.stockService.GetStockBySku(ctx, item.SKU)
		if err != nil {
			return nil, err
		}
		// Adjust stock inventory
		finalOnHand := stock.OnHand - item.Qty
		finalReserved := stock.Reserved - item.Qty
		req := apimodel.StockRequest{
			SKU:      stock.SKU,
			OnHand:   &finalOnHand,
			Reserved: &finalReserved,
		}
		// Save stock inventory
		_, err = ro.stockService.AdjustInventory(ctx, req, unit)
		if err != nil {
			return nil, err
		}
		// TODO: Deactivate reservation item
	}

	// Update reservation
	reservation, err = ro.reservationService.CommitReservation(ctx, reservation, params.OrderId, unit)
	if err != nil {
		return nil, err
	}

	err = unit.Commit()
	if err != nil {
		return nil, err
	}
	return reservation, nil
}
