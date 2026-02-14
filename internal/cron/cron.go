package cron

import (
	"context"
	"github.com/robfig/cron/v3"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/service"
)

// InitCrons initializes the cron jobs for the application.
func InitCrons(
	reservationService service.ReservationServiceInterface,
	logger *logger.Logger,
	config *config.Config,
) *cron.Cron {
	ctx := context.Background()
	c := cron.New()
	_, err := c.AddFunc(config.QuoteExpirationSettings.ExpireQuoteCronSpec, func() {
		processed, err := reservationService.ProcessExpiredReservations(ctx)
		if err != nil {
			logger.Errorf("expire cron error: %s", err)
		}
		logger.Infof("Expired reservations processed: %d", processed)
	})
	if err != nil {
		return nil
	}
	_, err = c.AddFunc(config.ArchiveSettings.ArchiveReservationsCronSpec, func() {

		deleted, err := reservationService.ArchiveReservations(ctx)
		if err != nil {
			logger.Errorf("cleanup cron error: %d", err)
		}
		logger.Infof("Archived reservations: %d", deleted)
	})
	if err != nil {
		return nil
	}

	logger.Info("Crons initialized successfully")
	return c
}
