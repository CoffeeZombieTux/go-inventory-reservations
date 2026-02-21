package application

import (
	"context"
	"go-inventory-reservations/internal/logger"
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
		ro.notifyExpiredReservationAsync(reservation)
	}
	return successCounter, failureCounter, nil
}

// processExpiredReservation expires a single reservation and releases its reserved stock.
func (ro *ReservationOrchestrator) processExpiredReservation(
	ctx context.Context,
	reservation *model.Reservation,
) error {
	err := uow.WithUnitOfWork(ctx, ro.uowManager, func(unit *uow.UnitOfWork) error {
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

// notifyExpiredReservationAsync asynchronously notifies the quote owner about the reservation expiration.
func (ro *ReservationOrchestrator) notifyExpiredReservationAsync(reservation *model.Reservation) {
	if ro.quoteNotifier == nil || reservation == nil {
		return
	}

	reservationCopy := *reservation
	go func(res model.Reservation) {
		notifyErr := ro.quoteNotifier.NotifyQuoteExpired(context.Background(), &res)
		if notifyErr != nil {
			ro.logger.WithError(notifyErr).WithFields(logger.Fields{
				"quote_id":       res.QuoteId,
				"reservation_id": res.ReservationId.String(),
				"status":         res.Status,
			}).Error(logger.LogMessageQuoteExpirationNotifyFailed)
		}
	}(reservationCopy)
}
