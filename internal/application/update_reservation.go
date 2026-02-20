package application

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/service"
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

	reservation, err := ro.reservationService.GetReservationById(ctx, *params.ReservationId)
	if err != nil {
		return nil, err
	}

	requestItemsHash := service.BuildReservationItemsHashFromRequests(params.Items)
	if reservation.QuoteId != nil &&
		*reservation.QuoteId == params.QuoteId &&
		itemsHashEqual(reservation.ItemsHash, requestItemsHash) {
		items, err := ro.reservationItemService.GetReservationItems(ctx, reservation.ReservationId, nil)
		if err != nil {
			return nil, err
		}
		result.Reservation = reservation
		result.Items = items
		return &result, nil
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

	// Update reservation
	reservation, err = ro.reservationService.UpdateReservation(ctx, params, unit)
	if err != nil {
		return nil, err
	}
	result.Reservation = reservation

	// Get reservation items
	reservedItems, err := ro.reservationItemService.GetReservationItems(ctx, reservation.ReservationId, unit)
	if err != nil {
		return nil, err
	}
	items := params.Items

	// Check if there are enough stock items for all items in the reservation
	for _, item := range items {
		stock, err := ro.stockService.GetStockBySku(ctx, item.SKU, unit)
		if err != nil {
			return nil, err
		}

		// Get quantity from request
		requiredQty := item.Quantity

		// Check if there are already reserved items for this SKU and this reservation
		// CASE 1: There are no reserved items for this SKU in this reservation [ok is false]
		// CASE 2: There are reserved items for this SKU in this reservation,
		//         but the reserved quantity is less than the quantity to be reserved
		//         In this case, we have to increase the reserved quantity of the SKU in the stock [requiredQty > 0]
		// CASE 3: There are reserved items for this SKU in this reservation,
		//         and the reserved quantity is more than the quantity to be reserved
		//         In this case, we have to decrease the reserved quantity of the SKU in the stock [requiredQty < 0]
		// CASE 4: There are reserved items for this SKU in this reservation,
		//         and the reserved quantity is equal to the quantity to be reserved
		//         In this case, we have to do nothing [item.Quantity == val.Qty]
		val, ok := reservedItems[item.SKU]
		if ok {
			requiredQty = item.Quantity - val.Qty
			if item.Quantity == val.Qty {
				result.Items[val.SKU] = val
				continue
			}
			if requiredQty < 0 {
				_, err := ro.stockService.ReserveStock(ctx, stock.SKU, requiredQty, unit)
				if err != nil {
					return nil, err
				}
			} else {
				if ro.stockService.CalculateAvailability(ctx, stock) < requiredQty {
					return nil, fmt.Errorf("insufficient stock for SKU %s", item.SKU)
				}
				_, err := ro.stockService.ReserveStock(ctx, stock.SKU, requiredQty, unit)
				if err != nil {
					return nil, err
				}
			}
			updatedItem, err := ro.reservationItemService.UpdateReservationItem(
				ctx,
				item,
				reservation.ReservationId,
				unit,
			)
			if err != nil {
				return nil, err
			}
			result.Items[item.SKU] = updatedItem
		} else {
			createdItem, err := ro.reservationItemService.CreateReservationItem(
				ctx,
				item,
				reservation.ReservationId,
				unit,
			)
			if err != nil {
				return nil, err
			}
			stock, err := ro.stockService.GetStockBySku(ctx, item.SKU, unit)
			if err != nil {
				return nil, err
			}
			if ro.stockService.CalculateAvailability(ctx, stock) < requiredQty {
				return nil, fmt.Errorf("insufficient stock for SKU %s", item.SKU)
			}
			_, err = ro.stockService.ReserveStock(ctx, stock.SKU, requiredQty, unit)
			if err != nil {
				return nil, err
			}
			result.Items[item.SKU] = createdItem
		}
	}

	err = unit.Commit()
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func itemsHashEqual(a *string, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
