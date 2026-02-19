package cron

import (
	"context"
	"go-inventory-reservations/internal/application"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/service"

	"github.com/robfig/cron/v3"
)

// InitCrons initializes the cron jobs for the application.
func InitCrons(
	reservationOrchestrator application.ReservationOrchestratorInterface,
	reservationService service.ReservationServiceInterface,
	logger *logger.Logger,
	config *config.Config,
) *cron.Cron {
	ctx := context.Background()
	c := cron.New()
	_, err := c.AddFunc(config.QuoteExpirationSettings.ExpireQuoteCronSpec, func() {
		success, fail, err := reservationOrchestrator.ProcessExpiredReservations(ctx)
		if err != nil {
			logger.Errorf("expire cron error: %s", err)
		}
		logger.Infof("Expired reservations processed cron finished. Success: %d, Fail: %d", success, fail)
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
