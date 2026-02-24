//go:build integration
// +build integration

package application

import (
	"context"
	"database/sql"
	"errors"
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

func TestProcessExpiredReservation_ReturnsVersionConflict(t *testing.T) {
	ctx := context.Background()
	db := testutil.OpenTestDB(t)
	defer db.Close()
	testutil.ResetAndMigrateTestDB(t, ctx, db)

	ro, stockSvc, resRepo, resSvc := newOrchestratorForExpiredTest(t, db)

	sku := "SKU-EXPIRE-" + uuid.NewString()
	onHand := 8
	reserved := 0
	_, err := stockSvc.CreateStock(ctx, apimodel.StockRequest{
		SKU:      sku,
		OnHand:   &onHand,
		Reserved: &reserved,
	})
	if err != nil {
		t.Fatalf("create stock failed: %v", err)
	}

	created, err := ro.CreateReservation(ctx, apimodel.CreateReservationRequest{
		QuoteId: "QUOTE-" + uuid.NewString(),
		Items: []apimodel.ReservationItemRequest{
			{SKU: sku, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("create reservation failed: %v", err)
	}

	stale, err := resSvc.GetReservationById(ctx, created.Reservation.ReservationId)
	if err != nil {
		t.Fatalf("get reservation failed: %v", err)
	}

	current, err := resRepo.GetById(ctx, created.Reservation.ReservationId)
	if err != nil {
		t.Fatalf("get reservation for concurrent update failed: %v", err)
	}
	current.QuoteId = "QUOTE-UPDATED-" + uuid.NewString()
	if _, err := resRepo.Save(ctx, current, nil); err != nil {
		t.Fatalf("concurrent update failed: %v", err)
	}

	err = ro.processExpiredReservation(ctx, stale)
	if err == nil {
		t.Fatalf("expected version conflict error")
	}
	if !errors.Is(err, repository.ErrReservationVersionConflict) {
		t.Fatalf("expected version conflict, got: %v", err)
	}

	stock, err := stockSvc.GetStockBySku(ctx, sku)
	if err != nil {
		t.Fatalf("get stock failed: %v", err)
	}
	if stock.Reserved != 2 {
		t.Fatalf("expected reserved stock to remain unchanged on conflict, got %d", stock.Reserved)
	}
}

func newOrchestratorForExpiredTest(
	t *testing.T,
	db *sql.DB,
) (
	*ReservationOrchestrator,
	service.StockServiceInterface,
	repository.ReservationRepositoryInterface,
	service.ReservationServiceInterface,
) {
	t.Helper()

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

	ro := &ReservationOrchestrator{
		uowManager:             uowManager,
		stockService:           stockSvc,
		reservationService:     resSvc,
		reservationItemService: resItemsSvc,
		quoteNotifier:          nil,
		logger:                 log,
	}

	return ro, stockSvc, resRepo, resSvc
}
