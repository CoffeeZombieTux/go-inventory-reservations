package application

import (
	"context"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"

	"github.com/google/uuid"
)

// ReleaseReservation orchestrates the process of releasing a reservation.
// It updates the stock inventory and updates the reservation status.
// It also deletes the reservation items.
func (ro *ReservationOrchestrator) ReleaseReservation(ctx context.Context, id uuid.UUID) error {
	reservation, err := ro.reservationService.GetReservationById(ctx, id)
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
	// Update reservation status
	err = ro.reservationService.ReleaseReservation(ctx, id, unit)
	if err != nil {
		return err
	}

	// Release reservation items
	err = ro.releaseReservationStocks(ctx, reservation.ReservationId, unit)
	if err != nil {
		return err
	}

	err = unit.Commit()
	if err != nil {
		return err
	}
	return nil
}

// releaseReservationStocks updates the stock inventory.
func (ro *ReservationOrchestrator) releaseReservationStocks(
	ctx context.Context,
	reservationId uuid.UUID,
	uow *uow.UnitOfWork,
) error {
	// Get reservation items
	reservedItems, err := ro.reservationItemService.GetReservationItems(ctx, reservationId, uow)
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
		finalReserved := stock.Reserved - item.Qty
		req := apimodel.StockRequest{
			SKU:      stock.SKU,
			OnHand:   nil,
			Reserved: &finalReserved,
		}

		// Save stock inventory
		_, err = ro.stockService.AdjustInventory(ctx, req, uow)
		if err != nil {
			return err
		}

		_, err = ro.reservationItemService.SetReservationItemActive(ctx, reservationId, item.SKU, false, uow)
		if err != nil {
			return err
		}
	}
	return nil
}
