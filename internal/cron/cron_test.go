package cron

import (
	"context"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"
	"testing"

	"github.com/google/uuid"
)

type cronReservationOrchestratorStub struct {
	processCalls int
}

func (c *cronReservationOrchestratorStub) CreateReservation(ctx context.Context, params apimodel.CreateReservationRequest) (*apimodel.ReservationResponse, error) {
	panic("not used")
}
func (c *cronReservationOrchestratorStub) UpdateReservation(ctx context.Context, params apimodel.UpdateReservationRequest) (*apimodel.ReservationResponse, error) {
	panic("not used")
}
func (c *cronReservationOrchestratorStub) GetReservationById(ctx context.Context, reservationId uuid.UUID) (*apimodel.ReservationResponse, error) {
	panic("not used")
}
func (c *cronReservationOrchestratorStub) GetReservationByQuoteId(ctx context.Context, quoteId string) (*apimodel.ReservationResponse, error) {
	panic("not used")
}
func (c *cronReservationOrchestratorStub) GetReservationByOrderId(ctx context.Context, orderId string) (*apimodel.ReservationResponse, error) {
	panic("not used")
}
func (c *cronReservationOrchestratorStub) CommitReservation(ctx context.Context, params apimodel.CommitReservationRequest) (*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationOrchestratorStub) ReleaseReservation(ctx context.Context, id uuid.UUID) error {
	panic("not used")
}
func (c *cronReservationOrchestratorStub) RevertReservation(ctx context.Context, request apimodel.RevertReservationRequest) error {
	panic("not used")
}
func (c *cronReservationOrchestratorStub) ProcessExpiredReservations(ctx context.Context) (successCounter int, failureCounter int, err error) {
	c.processCalls++
	return 1, 0, nil
}

type cronReservationServiceStub struct {
	archiveCalls int
}

func (c *cronReservationServiceStub) GetReservationById(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationServiceStub) GetReservationByIdForUpdate(ctx context.Context, id uuid.UUID, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationServiceStub) GetReservationByQuoteId(ctx context.Context, quoteId string) (*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationServiceStub) GetReservationByOrderId(ctx context.Context, orderId string) (*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationServiceStub) GetToBeExpiredReservations(ctx context.Context) ([]*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationServiceStub) AttachOrder(ctx context.Context, request apimodel.AttachOrderRequest) error {
	panic("not used")
}
func (c *cronReservationServiceStub) ArchiveReservations(ctx context.Context) (int, error) {
	c.archiveCalls++
	return 2, nil
}
func (c *cronReservationServiceStub) CreateReservationHelper(ctx context.Context, request apimodel.CreateReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationServiceStub) UpdateReservationHelper(ctx context.Context, request apimodel.UpdateReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationServiceStub) CommitReservationHelper(ctx context.Context, request apimodel.CommitReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationServiceStub) ReleaseReservationHelper(ctx context.Context, id uuid.UUID, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationServiceStub) RevertReservationHelper(ctx context.Context, request apimodel.RevertReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (c *cronReservationServiceStub) ExpireReservationHelper(ctx context.Context, reservation *model.Reservation, u *uow.UnitOfWork) error {
	panic("not used")
}

func TestInitCronsAndRunJobs(t *testing.T) {
	ro := &cronReservationOrchestratorStub{}
	rs := &cronReservationServiceStub{}
	cfg := &config.Config{
		QuoteExpirationSettings: config.QuoteExpirationSettings{
			ExpireQuoteCronSpec: "* * * * *",
		},
		ArchiveSettings: config.ArchiveSettings{
			ArchiveReservationsCronSpec: "* * * * *",
		},
	}

	c := InitCrons(ro, rs, logger.New("error", "text"), cfg)
	entries := c.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 cron entries, got %d", len(entries))
	}

	for _, entry := range entries {
		entry.Job.Run()
	}

	if ro.processCalls == 0 {
		t.Fatalf("expected expire cron job to run")
	}
	if rs.archiveCalls == 0 {
		t.Fatalf("expected archive cron job to run")
	}
}
