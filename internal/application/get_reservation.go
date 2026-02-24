package application

import (
	"context"
	"github.com/google/uuid"
	apimodel "go-inventory-reservations/internal/model/api"
)

// GetReservationById retrieves a reservation by its ID with items included
func (ro *ReservationOrchestrator) GetReservationById(
	ctx context.Context,
	reservationId uuid.UUID,
) (*apimodel.ReservationResponse, error) {
	reservation, err := ro.reservationService.GetReservationById(ctx, reservationId)
	if err != nil {
		return nil, err
	}
	items, err := ro.reservationItemService.GetReservationItems(ctx, reservation.ReservationId, nil)
	if err != nil {
		return nil, err
	}
	return &apimodel.ReservationResponse{
		Reservation: reservation,
		Items:       items,
	}, nil
}

// GetReservationByQuoteId retrieves a reservation by its quote ID with items included
func (ro *ReservationOrchestrator) GetReservationByQuoteId(
	ctx context.Context,
	quoteId string,
) (*apimodel.ReservationResponse, error) {
	reservation, err := ro.reservationService.GetReservationByQuoteId(ctx, quoteId)
	if err != nil {
		return nil, err
	}
	items, err := ro.reservationItemService.GetReservationItems(ctx, reservation.ReservationId, nil)
	if err != nil {
		return nil, err
	}
	return &apimodel.ReservationResponse{
		Reservation: reservation,
		Items:       items,
	}, nil
}

// GetReservationByOrderId retrieves a reservation by its order ID with items included
func (ro *ReservationOrchestrator) GetReservationByOrderId(
	ctx context.Context,
	orderId string,
) (*apimodel.ReservationResponse, error) {
	reservation, err := ro.reservationService.GetReservationByOrderId(ctx, orderId)
	if err != nil {
		return nil, err
	}
	items, err := ro.reservationItemService.GetReservationItems(ctx, reservation.ReservationId, nil)
	return &apimodel.ReservationResponse{
		Reservation: reservation,
		Items:       items,
	}, err
}
