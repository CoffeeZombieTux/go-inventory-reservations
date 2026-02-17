package application

import (
	"context"
	apimodel "go-inventory-reservations/internal/model/api"
)

// RevertReservation orchestrates the process of reverting a reservation.
// It updates the stock inventory and updates the reservation status.
func (ro *ReservationOrchestrator) RevertReservation(
	ctx context.Context,
	request apimodel.RevertReservationRequest,
) error {
	reservation, err := ro.reservationService.GetReservationById(ctx, *request.ReservationId)
	if err != nil {
		return err
	}

	unit, err := ro.uowManager.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			unit.Rollback()
		}
	}()

	// Get reservation items
	reservedItems, err := ro.reservationItemService.GetReservationItems(ctx, reservation.ReservationId, unit)
	if err != nil {
		return err
	}

	for _, item := range reservedItems {
		// Get stock for item
		stock, err := ro.stockService.GetStockBySku(ctx, item.SKU)
		if err != nil {
			return err
		}

		// Adjust stock inventory
		finalOnHand := stock.OnHand + item.Qty
		req := apimodel.StockRequest{
			SKU:      stock.SKU,
			OnHand:   &finalOnHand,
			Reserved: nil,
		}

		// Save stock inventory
		_, err = ro.stockService.AdjustInventory(ctx, req, unit)
		if err != nil {
			return err
		}
	}

	// Update reservation
	err = ro.reservationService.RevertReservation(ctx, request, unit)
	if err != nil {
		return err
	}

	err = unit.Commit()
	if err != nil {
		return err
	}
	return nil
}
