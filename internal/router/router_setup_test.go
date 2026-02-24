package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-inventory-reservations/internal/application"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/handler"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type routerStockServiceStub struct{}

func (s *routerStockServiceStub) CreateStock(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error) {
	return &model.Stock{SKU: req.SKU}, nil
}
func (s *routerStockServiceStub) GetStockBySku(ctx context.Context, sku string) (*apimodel.StockResponse, error) {
	return &apimodel.StockResponse{SKU: sku, OnHand: 1, Reserved: 0, Available: 1}, nil
}
func (s *routerStockServiceStub) GetStockBySkuForUpdate(ctx context.Context, sku string, u *uow.UnitOfWork) (*model.Stock, error) {
	return &model.Stock{SKU: sku, OnHand: 1, Reserved: 0}, nil
}
func (s *routerStockServiceStub) GetStocks(ctx context.Context, requestedLimit, requestedOffset int) ([]*apimodel.StockResponse, *apimodel.PaginationResponse, string, error) {
	return []*apimodel.StockResponse{}, &apimodel.PaginationResponse{Limit: 50, Offset: 0}, "ok", nil
}
func (s *routerStockServiceStub) ReserveStock(ctx context.Context, sku string, qty int, u *uow.UnitOfWork) (*model.Stock, error) {
	return &model.Stock{SKU: sku}, nil
}
func (s *routerStockServiceStub) AdjustInventory(ctx context.Context, req apimodel.StockRequest, u *uow.UnitOfWork) (*model.Stock, error) {
	return &model.Stock{SKU: req.SKU}, nil
}
func (s *routerStockServiceStub) DeleteStock(ctx context.Context, sku string) error { return nil }
func (s *routerStockServiceStub) CalculateAvailability(ctx context.Context, stock *model.Stock) int {
	return stock.OnHand - stock.Reserved
}

type routerReservationOrchestratorStub struct{}

func (r *routerReservationOrchestratorStub) CreateReservation(ctx context.Context, params apimodel.CreateReservationRequest) (*apimodel.ReservationResponse, error) {
	id := uuid.New()
	return &apimodel.ReservationResponse{Reservation: &model.Reservation{ReservationId: id}, Items: map[string]*model.ReservationItem{}}, nil
}
func (r *routerReservationOrchestratorStub) UpdateReservation(ctx context.Context, params apimodel.UpdateReservationRequest) (*apimodel.ReservationResponse, error) {
	id := uuid.New()
	return &apimodel.ReservationResponse{Reservation: &model.Reservation{ReservationId: id}, Items: map[string]*model.ReservationItem{}}, nil
}
func (r *routerReservationOrchestratorStub) GetReservationById(ctx context.Context, reservationId uuid.UUID) (*apimodel.ReservationResponse, error) {
	return &apimodel.ReservationResponse{Reservation: &model.Reservation{ReservationId: reservationId}, Items: map[string]*model.ReservationItem{}}, nil
}
func (r *routerReservationOrchestratorStub) GetReservationByQuoteId(ctx context.Context, quoteId string) (*apimodel.ReservationResponse, error) {
	id := uuid.New()
	return &apimodel.ReservationResponse{Reservation: &model.Reservation{ReservationId: id, QuoteId: quoteId}, Items: map[string]*model.ReservationItem{}}, nil
}
func (r *routerReservationOrchestratorStub) GetReservationByOrderId(ctx context.Context, orderId string) (*apimodel.ReservationResponse, error) {
	id := uuid.New()
	return &apimodel.ReservationResponse{Reservation: &model.Reservation{ReservationId: id}, Items: map[string]*model.ReservationItem{}}, nil
}
func (r *routerReservationOrchestratorStub) CommitReservation(ctx context.Context, params apimodel.CommitReservationRequest) (*model.Reservation, error) {
	return &model.Reservation{ReservationId: *params.ReservationId}, nil
}
func (r *routerReservationOrchestratorStub) ReleaseReservation(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (r *routerReservationOrchestratorStub) RevertReservation(ctx context.Context, request apimodel.RevertReservationRequest) error {
	return nil
}
func (r *routerReservationOrchestratorStub) ProcessExpiredReservations(ctx context.Context) (successCounter int, failureCounter int, err error) {
	return 0, 0, nil
}

type routerAdminOrchestratorStub struct{}

func (a *routerAdminOrchestratorStub) DeleteStock(ctx context.Context, sku string) error { return nil }
func (a *routerAdminOrchestratorStub) AdjustInventory(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error) {
	return &model.Stock{SKU: req.SKU}, nil
}
func (a *routerAdminOrchestratorStub) GetActiveReservationItemsBySku(ctx context.Context, sku string, requestedLimit int, requestedOffset int) ([]*model.ReservationItem, *apimodel.PaginationResponse, string, error) {
	return []*model.ReservationItem{}, &apimodel.PaginationResponse{Limit: 50, Offset: 0}, "ok", nil
}

type routerReservationServiceStub struct{}

func (s *routerReservationServiceStub) GetReservationById(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	return &model.Reservation{ReservationId: id}, nil
}
func (s *routerReservationServiceStub) GetReservationByIdForUpdate(ctx context.Context, id uuid.UUID, u *uow.UnitOfWork) (*model.Reservation, error) {
	return &model.Reservation{ReservationId: id}, nil
}
func (s *routerReservationServiceStub) GetReservationByQuoteId(ctx context.Context, quoteId string) (*model.Reservation, error) {
	return &model.Reservation{ReservationId: uuid.New()}, nil
}
func (s *routerReservationServiceStub) GetReservationByOrderId(ctx context.Context, orderId string) (*model.Reservation, error) {
	return &model.Reservation{ReservationId: uuid.New()}, nil
}
func (s *routerReservationServiceStub) GetToBeExpiredReservations(ctx context.Context) ([]*model.Reservation, error) {
	return []*model.Reservation{}, nil
}
func (s *routerReservationServiceStub) AttachOrder(ctx context.Context, request apimodel.AttachOrderRequest) error {
	return nil
}
func (s *routerReservationServiceStub) ArchiveReservations(ctx context.Context) (int, error) {
	return 0, nil
}
func (s *routerReservationServiceStub) CreateReservationHelper(ctx context.Context, request apimodel.CreateReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	return &model.Reservation{ReservationId: uuid.New()}, nil
}
func (s *routerReservationServiceStub) UpdateReservationHelper(ctx context.Context, request apimodel.UpdateReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	return &model.Reservation{ReservationId: *request.ReservationId}, nil
}
func (s *routerReservationServiceStub) CommitReservationHelper(ctx context.Context, request apimodel.CommitReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	return &model.Reservation{ReservationId: *request.ReservationId}, nil
}
func (s *routerReservationServiceStub) ReleaseReservationHelper(ctx context.Context, id uuid.UUID, u *uow.UnitOfWork) (*model.Reservation, error) {
	return &model.Reservation{ReservationId: id}, nil
}
func (s *routerReservationServiceStub) RevertReservationHelper(ctx context.Context, request apimodel.RevertReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	return &model.Reservation{ReservationId: *request.ReservationId}, nil
}
func (s *routerReservationServiceStub) ExpireReservationHelper(ctx context.Context, reservation *model.Reservation, u *uow.UnitOfWork) error {
	return nil
}

func TestSetupRoutesAppliesAuthAndHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handlersPool := handler.NewHandlersPool(
		&routerReservationOrchestratorStub{},
		&routerAdminOrchestratorStub{},
		&routerStockServiceStub{},
		&routerReservationServiceStub{},
		logger.New("error", "text"),
	)
	cfg := &config.Config{
		Auth: config.AuthConfig{
			AdminToken:  "admin-token",
			PublicToken: "public-token",
		},
	}

	SetupRoutes(engine, *handlersPool, cfg, logger.New("error", "text"))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected ping 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/stock/SKU-1", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected stock unauthorized without token, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/stock/SKU-1", nil)
	req.Header.Set("Authorization", "Bearer public-token")
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected stock 200 with public token, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/admin/stock/SKU-1", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected admin delete 200 with admin token, got %d", rec.Code)
	}
}

var (
	_ application.ReservationOrchestratorInterface = &routerReservationOrchestratorStub{}
	_ application.AdminStockOrchestratorInterface  = &routerAdminOrchestratorStub{}
)
