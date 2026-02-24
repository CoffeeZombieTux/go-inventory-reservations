//go:build integration
// +build integration

package application_test

import (
	"context"
	"database/sql"
	"go-inventory-reservations/internal/application"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/logger"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/repository"
	"go-inventory-reservations/internal/service"
	"go-inventory-reservations/internal/testutil"
	"go-inventory-reservations/internal/uow"
	"testing"

	"github.com/google/uuid"
)

type flowDeps struct {
	orch        application.ReservationOrchestratorInterface
	stockSvc    service.StockServiceInterface
	resSvc      service.ReservationServiceInterface
	resItemsSvc service.ReservationItemsServiceInterface
}

func TestFlow_CreateAttachCommitReservation(t *testing.T) {
	ctx := context.Background()
	db := testutil.OpenTestDB(t)
	defer db.Close()
	testutil.ResetAndMigrateTestDB(t, ctx, db)

	deps := newFlowDeps(db)
	sku := "SKU-COMMIT-" + uuid.NewString()
	quoteID := "QUOTE-" + uuid.NewString()
	orderID := "ORDER-" + uuid.NewString()

	onHand := 10
	reserved := 0
	_, err := deps.stockSvc.CreateStock(ctx, apimodel.StockRequest{
		SKU:      sku,
		OnHand:   &onHand,
		Reserved: &reserved,
	})
	if err != nil {
		t.Fatalf("create stock failed: %v", err)
	}

	created, err := deps.orch.CreateReservation(ctx, apimodel.CreateReservationRequest{
		QuoteId: quoteID,
		Items: []apimodel.ReservationItemRequest{
			{SKU: sku, Quantity: 3},
		},
	})
	if err != nil {
		t.Fatalf("create reservation failed: %v", err)
	}

	if created.Reservation.Status != "PENDING" {
		t.Fatalf("expected PENDING status, got %s", created.Reservation.Status)
	}

	err = deps.resSvc.AttachOrder(ctx, apimodel.AttachOrderRequest{
		ReservationId: &created.Reservation.ReservationId,
		OrderId:       orderID,
	})
	if err != nil {
		t.Fatalf("attach order failed: %v", err)
	}

	_, err = deps.orch.CommitReservation(ctx, apimodel.CommitReservationRequest{
		ReservationId: &created.Reservation.ReservationId,
		OrderId:       orderID,
	})
	if err != nil {
		t.Fatalf("commit reservation failed: %v", err)
	}

	stock, err := deps.stockSvc.GetStockBySku(ctx, sku)
	if err != nil {
		t.Fatalf("get stock failed: %v", err)
	}
	if stock.OnHand != 7 || stock.Reserved != 0 || stock.Available != 7 {
		t.Fatalf("unexpected stock state after commit: %+v", stock)
	}

	reservation, err := deps.resSvc.GetReservationById(ctx, created.Reservation.ReservationId)
	if err != nil {
		t.Fatalf("get reservation failed: %v", err)
	}
	if reservation.Status != "COMMITTED" {
		t.Fatalf("expected COMMITTED status, got %s", reservation.Status)
	}
}

func TestFlow_CreateAndReleaseReservation(t *testing.T) {
	ctx := context.Background()
	db := testutil.OpenTestDB(t)
	defer db.Close()
	testutil.ResetAndMigrateTestDB(t, ctx, db)

	deps := newFlowDeps(db)
	sku := "SKU-RELEASE-" + uuid.NewString()
	quoteID := "QUOTE-" + uuid.NewString()

	onHand := 9
	reserved := 0
	_, err := deps.stockSvc.CreateStock(ctx, apimodel.StockRequest{
		SKU:      sku,
		OnHand:   &onHand,
		Reserved: &reserved,
	})
	if err != nil {
		t.Fatalf("create stock failed: %v", err)
	}

	created, err := deps.orch.CreateReservation(ctx, apimodel.CreateReservationRequest{
		QuoteId: quoteID,
		Items: []apimodel.ReservationItemRequest{
			{SKU: sku, Quantity: 4},
		},
	})
	if err != nil {
		t.Fatalf("create reservation failed: %v", err)
	}

	err = deps.orch.ReleaseReservation(ctx, created.Reservation.ReservationId)
	if err != nil {
		t.Fatalf("release reservation failed: %v", err)
	}

	stock, err := deps.stockSvc.GetStockBySku(ctx, sku)
	if err != nil {
		t.Fatalf("get stock failed: %v", err)
	}
	if stock.OnHand != 9 || stock.Reserved != 0 || stock.Available != 9 {
		t.Fatalf("unexpected stock state after release: %+v", stock)
	}

	reservation, err := deps.resSvc.GetReservationById(ctx, created.Reservation.ReservationId)
	if err != nil {
		t.Fatalf("get reservation failed: %v", err)
	}
	if reservation.Status != "RELEASED" {
		t.Fatalf("expected RELEASED status, got %s", reservation.Status)
	}

	items, err := deps.resItemsSvc.GetReservationItems(ctx, created.Reservation.ReservationId, nil)
	if err != nil {
		t.Fatalf("get reservation items failed: %v", err)
	}
	item, ok := items[sku]
	if !ok {
		t.Fatalf("expected reservation item for %s", sku)
	}
	if item.IsActive {
		t.Fatalf("expected reservation item to be inactive after release")
	}
}

func newFlowDeps(db *sql.DB) flowDeps {
	log := logger.New("error", "text")
	cfg := &config.Config{
		QuoteExpirationSettings: config.QuoteExpirationSettings{
			QuoteExpirationMinutes: 15,
			Limit:                  1000,
		},
		ArchiveSettings: config.ArchiveSettings{
			ArchiveReservationsAfterDays: 30,
			Limit:                        10000,
		},
		ReservationItemSettings: config.ReservationItemSettings{
			MaxQuantity: 20,
		},
	}

	uowManager := uow.NewUnitOfWorkManager(db)
	stockRepo := repository.NewStockRepository(db, log)
	resRepo := repository.NewReservationRepository(db, log)
	resItemRepo := repository.NewReservationItemsRepository(db, log)

	stockSvc := service.NewStockService(stockRepo)
	resSvc := service.NewReservationService(resRepo, cfg)
	resItemsSvc := service.NewReservationItemsService(resItemRepo, cfg)
	orch := application.NewReservationOrchestrator(
		uowManager,
		stockSvc,
		resSvc,
		resItemsSvc,
		nil,
		log,
	)

	return flowDeps{
		orch:        orch,
		stockSvc:    stockSvc,
		resSvc:      resSvc,
		resItemsSvc: resItemsSvc,
	}
}
