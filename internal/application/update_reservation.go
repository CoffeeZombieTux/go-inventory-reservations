package application

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"

	"github.com/google/uuid"
)

// UpdateReservation orchestrates the process of updating a reservation.
// It updates the reservation and reservation items.
// It also updates the stock inventory.
// It also updates the reservation status if the reservation is committed.
func (ro *ReservationOrchestrator) UpdateReservation(
	ctx context.Context,
	params apimodel.UpdateReservationRequest,
) (*apimodel.ReservationResponse, error) {
	result := apimodel.ReservationResponse{Items: make(map[string]*model.ReservationItem)}

	err := ro.withUnitOfWork(ctx, func(unit *uow.UnitOfWork) error {
		// Update reservation
		reservation, err := ro.reservationService.UpdateReservationHelper(ctx, params, unit)
		if err != nil {
			return err
		}
		result.Reservation = reservation

		// Get reservation items
		reservedItems, err := ro.reservationItemService.GetReservationItems(ctx, reservation.ReservationId, unit)
		if err != nil {
			return err
		}

		return ro.syncUpdatedReservationItems(
			ctx,
			reservation.ReservationId,
			params.Items,
			reservedItems,
			result.Items,
			unit,
		)
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// syncUpdatedReservationItems updates/creates requested items and removes items absent in the request.
func (ro *ReservationOrchestrator) syncUpdatedReservationItems(
	ctx context.Context,
	reservationID uuid.UUID,
	items []apimodel.ReservationItemRequest,
	reservedItems map[string]*model.ReservationItem,
	resultItems map[string]*model.ReservationItem,
	unit *uow.UnitOfWork,
) error {
	for _, item := range items {
		stock, err := ro.stockService.GetStockBySkuForUpdate(ctx, item.SKU, unit)
		if err != nil {
			return err
		}

		if existing, ok := reservedItems[item.SKU]; ok {
			delete(reservedItems, item.SKU)
			err = ro.applyExistingReservationItemUpdate(ctx, reservationID, item, existing, stock, resultItems, unit)
			if err != nil {
				return err
			}
			continue
		}

		err = ro.applyNewReservationItem(ctx, reservationID, item, resultItems, unit)
		if err != nil {
			return err
		}
	}

	return ro.removeMissingReservationItems(ctx, reservedItems, unit)
}

// applyExistingReservationItemUpdate updates an existing reservation item and adjusts reserved stock.
func (ro *ReservationOrchestrator) applyExistingReservationItemUpdate(
	ctx context.Context,
	reservationID uuid.UUID,
	item apimodel.ReservationItemRequest,
	existing *model.ReservationItem,
	stock *model.Stock,
	resultItems map[string]*model.ReservationItem,
	unit *uow.UnitOfWork,
) error {
	requiredQty := item.Quantity - existing.Qty
	if requiredQty == 0 {
		resultItems[existing.SKU] = existing
		return nil
	}

	updatedStock, err := ro.adjustReservedStockForUpdate(ctx, item.SKU, requiredQty, stock, unit)
	if err != nil {
		return err
	}

	if updatedStock.Reserved == 0 {
		return ro.reservationItemService.DeleteReservationItem(ctx, reservationID, item.SKU, unit)
	}

	updatedItem, err := ro.reservationItemService.UpdateReservationItem(ctx, item, reservationID, unit)
	if err != nil {
		return err
	}
	resultItems[item.SKU] = updatedItem
	return nil
}

// adjustReservedStockForUpdate applies stock reservation delta for reservation item updates.
func (ro *ReservationOrchestrator) adjustReservedStockForUpdate(
	ctx context.Context,
	sku string,
	requiredQty int,
	stock *model.Stock,
	unit *uow.UnitOfWork,
) (*model.Stock, error) {
	if requiredQty > 0 && ro.stockService.CalculateAvailability(ctx, stock) < requiredQty {
		return nil, fmt.Errorf("insufficient stock for SKU %s", sku)
	}

	return ro.stockService.ReserveStock(ctx, stock.SKU, requiredQty, unit)
}

// applyNewReservationItem creates a new reservation item and reserves the requested stock.
func (ro *ReservationOrchestrator) applyNewReservationItem(
	ctx context.Context,
	reservationID uuid.UUID,
	item apimodel.ReservationItemRequest,
	resultItems map[string]*model.ReservationItem,
	unit *uow.UnitOfWork,
) error {
	createdItem, err := ro.reservationItemService.CreateReservationItem(ctx, item, reservationID, unit)
	if err != nil {
		return err
	}

	stock, err := ro.stockService.GetStockBySkuForUpdate(ctx, item.SKU, unit)
	if err != nil {
		return err
	}

	if ro.stockService.CalculateAvailability(ctx, stock) < item.Quantity {
		return fmt.Errorf("insufficient stock for SKU %s", item.SKU)
	}

	_, err = ro.stockService.ReserveStock(ctx, stock.SKU, item.Quantity, unit)
	if err != nil {
		return err
	}

	resultItems[item.SKU] = createdItem
	return nil
}

// removeMissingReservationItems deletes reservation items not present in the update request and releases stock.
func (ro *ReservationOrchestrator) removeMissingReservationItems(
	ctx context.Context,
	reservedItems map[string]*model.ReservationItem,
	unit *uow.UnitOfWork,
) error {
	for _, restItem := range reservedItems {
		err := ro.reservationItemService.DeleteReservationItem(ctx, restItem.ReservationId, restItem.SKU, unit)
		if err != nil {
			return err
		}

		_, err = ro.stockService.ReserveStock(ctx, restItem.SKU, -restItem.Qty, unit)
		if err != nil {
			return err
		}
	}
	return nil
}
