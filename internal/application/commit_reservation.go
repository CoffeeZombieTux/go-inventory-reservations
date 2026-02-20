package application

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"

	"github.com/google/uuid"
)

// CommitReservation orchestrates the process of committing a reservation.
// It updates the stock inventory and updates the reservation status.
func (ro *ReservationOrchestrator) CommitReservation(
	ctx context.Context,
	params apimodel.CommitReservationRequest,
) (*model.Reservation, error) {
	var reservation *model.Reservation
	err := ro.withUnitOfWork(ctx, func(unit *uow.UnitOfWork) error {
		// Update reservation
		var err error
		reservation, err = ro.reservationService.CommitReservationHelper(ctx, params, unit)
		if err != nil {
			return err
		}

		reservedItems, err := ro.getReservationItemsForCommit(ctx, reservation.ReservationId, unit)
		if err != nil {
			return err
		}

		return ro.commitReservationItems(ctx, reservation.ReservationId, reservedItems, unit)
	})
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// getReservationItemsForCommit loads reservation items and ensures commit preconditions.
func (ro *ReservationOrchestrator) getReservationItemsForCommit(
	ctx context.Context,
	reservationID uuid.UUID,
	unit *uow.UnitOfWork,
) (map[string]*model.ReservationItem, error) {
	reservedItems, err := ro.reservationItemService.GetReservationItems(ctx, reservationID, unit)
	if err != nil {
		return nil, err
	}

	if len(reservedItems) == 0 {
		return nil, fmt.Errorf(
			"you cannot commit reservation %s because it has not been reserved items",
			reservationID.String(),
		)
	}

	return reservedItems, nil
}

// commitReservationItems applies stock movement and deactivates items for reservation commit.
func (ro *ReservationOrchestrator) commitReservationItems(
	ctx context.Context,
	reservationID uuid.UUID,
	reservedItems map[string]*model.ReservationItem,
	unit *uow.UnitOfWork,
) error {
	for _, item := range reservedItems {
		err := ro.commitReservationItem(ctx, reservationID, item, unit)
		if err != nil {
			return err
		}
	}
	return nil
}

// commitReservationItem applies stock movement and deactivates a single reservation item.
func (ro *ReservationOrchestrator) commitReservationItem(
	ctx context.Context,
	reservationID uuid.UUID,
	item *model.ReservationItem,
	unit *uow.UnitOfWork,
) error {
	stock, err := ro.stockService.GetStockBySkuForUpdate(ctx, item.SKU, unit)
	if err != nil {
		return err
	}

	finalOnHand := stock.OnHand - item.Qty
	finalReserved := stock.Reserved - item.Qty
	req := apimodel.StockRequest{
		SKU:      stock.SKU,
		OnHand:   &finalOnHand,
		Reserved: &finalReserved,
	}

	_, err = ro.stockService.AdjustInventory(ctx, req, unit)
	if err != nil {
		return err
	}

	_, err = ro.reservationItemService.SetReservationItemActive(ctx, reservationID, item.SKU, false, unit)
	if err != nil {
		return err
	}

	return nil
}
