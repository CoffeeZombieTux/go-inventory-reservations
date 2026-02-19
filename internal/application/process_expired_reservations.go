package application

import (
	"context"
	"go-inventory-reservations/internal/model"
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

func (ro *ReservationOrchestrator) processExpiredReservation(
	ctx context.Context,
	reservation *model.Reservation,
) error {
	unit, err := ro.uowManager.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = unit.Rollback()
		}
	}()

	err = ro.reservationService.ExpireReservation(ctx, reservation, unit)
	if err != nil {
		return err
	}

	err = ro.releaseReservationStocks(ctx, reservation.ReservationId, unit)
	if err != nil {
		return err
	}

	err = unit.Commit()
	return err
}
