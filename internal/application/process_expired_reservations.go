package application

import (
	"context"
	"go-inventory-reservations/internal/model"
	"go-inventory-reservations/internal/service"
	"go-inventory-reservations/internal/uow"
)

// ProcessExpiredReservations processes all expired reservations with items.
func (ro *ReservationOrchestrator) ProcessExpiredReservations(
	ctx context.Context,
) (successCounter int, failureCounter int, err error) {
	reservations, err := ro.reservationService.GetToBeExpiredReservations(ctx)
	if err != nil {
		return 0, 0, err
	}
	for _, reservation := range reservations {
		processErr := ro.processExpiredReservation(ctx, reservation)
		if processErr != nil {
			failureCounter++
			continue
		}
		successCounter++
	}
	return successCounter, failureCounter, nil
}

// processExpiredReservation expires a single reservation and releases its reserved stock.
func (ro *ReservationOrchestrator) processExpiredReservation(
	ctx context.Context,
	reservation *model.Reservation,
) error {
	err := ro.withUnitOfWork(ctx, func(unit *uow.UnitOfWork) error {
		err := ro.reservationService.ExpireReservationHelper(ctx, reservation, unit)
		if err != nil {
			if service.IsReservationVersionConflict(err) {
				return nil
			}
			return err
		}

		return ro.releaseReservationStocks(ctx, reservation.ReservationId, unit)
	})
	return err
}
