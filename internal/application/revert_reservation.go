package application

import (
	"context"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"
)

// RevertReservation orchestrates the process of reverting a reservation.
// It used for refund order and return products to stock inventory.
func (ro *ReservationOrchestrator) RevertReservation(
	ctx context.Context,
	request apimodel.RevertReservationRequest,
) error {
	return ro.withUnitOfWork(ctx, func(unit *uow.UnitOfWork) error {
		// Update reservation
		reservation, err := ro.reservationService.RevertReservationHelper(ctx, request, unit)
		if err != nil {
			return err
		}

		// Get reservation items
		reservedItems, err := ro.reservationItemService.GetReservationItems(ctx, reservation.ReservationId, unit)
		if err != nil {
			return err
		}

		for _, item := range reservedItems {
			// Get stock for item
			stock, err := ro.stockService.GetStockBySkuForUpdate(ctx, item.SKU, unit)
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
		return nil
	})
}
