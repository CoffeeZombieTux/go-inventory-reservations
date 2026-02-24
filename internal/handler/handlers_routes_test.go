package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-inventory-reservations/internal/apperror"
	"go-inventory-reservations/internal/application"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handlerStockServiceStub struct {
	getBySkuResp *apimodel.StockResponse
	getBySkuErr  error

	getStocksResp       []*apimodel.StockResponse
	getStocksPagination *apimodel.PaginationResponse
	getStocksMessage    string
	getStocksErr        error

	createResp *model.Stock
	createErr  error
}

func (s *handlerStockServiceStub) CreateStock(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error) {
	return s.createResp, s.createErr
}
func (s *handlerStockServiceStub) GetStockBySku(ctx context.Context, sku string) (*apimodel.StockResponse, error) {
	return s.getBySkuResp, s.getBySkuErr
}
func (s *handlerStockServiceStub) GetStockBySkuForUpdate(ctx context.Context, sku string, u *uow.UnitOfWork) (*model.Stock, error) {
	panic("not used")
}
func (s *handlerStockServiceStub) GetStocks(
	ctx context.Context,
	requestedLimit,
	requestedOffset int,
) ([]*apimodel.StockResponse, *apimodel.PaginationResponse, string, error) {
	return s.getStocksResp, s.getStocksPagination, s.getStocksMessage, s.getStocksErr
}
func (s *handlerStockServiceStub) ReserveStock(ctx context.Context, sku string, qty int, u *uow.UnitOfWork) (*model.Stock, error) {
	panic("not used")
}
func (s *handlerStockServiceStub) AdjustInventory(ctx context.Context, req apimodel.StockRequest, u *uow.UnitOfWork) (*model.Stock, error) {
	panic("not used")
}
func (s *handlerStockServiceStub) DeleteStock(ctx context.Context, sku string) error {
	panic("not used")
}
func (s *handlerStockServiceStub) CalculateAvailability(ctx context.Context, stock *model.Stock) int {
	panic("not used")
}

type handlerAdminOrchestratorStub struct {
	deleteErr   error
	adjustResp  *model.Stock
	adjustErr   error
	activeItems []*model.ReservationItem
	activePage  *apimodel.PaginationResponse
	activeMsg   string
	activeErr   error
}

func (a *handlerAdminOrchestratorStub) DeleteStock(ctx context.Context, sku string) error {
	return a.deleteErr
}
func (a *handlerAdminOrchestratorStub) AdjustInventory(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error) {
	return a.adjustResp, a.adjustErr
}
func (a *handlerAdminOrchestratorStub) GetActiveReservationItemsBySku(
	ctx context.Context,
	sku string,
	requestedLimit int,
	requestedOffset int,
) ([]*model.ReservationItem, *apimodel.PaginationResponse, string, error) {
	return a.activeItems, a.activePage, a.activeMsg, a.activeErr
}

type handlerReservationOrchestratorStub struct {
	getByIDResp    *apimodel.ReservationResponse
	getByIDErr     error
	getByQuoteResp *apimodel.ReservationResponse
	getByQuoteErr  error
}

func (r *handlerReservationOrchestratorStub) CreateReservation(ctx context.Context, params apimodel.CreateReservationRequest) (*apimodel.ReservationResponse, error) {
	panic("not used")
}
func (r *handlerReservationOrchestratorStub) UpdateReservation(ctx context.Context, params apimodel.UpdateReservationRequest) (*apimodel.ReservationResponse, error) {
	panic("not used")
}
func (r *handlerReservationOrchestratorStub) GetReservationById(ctx context.Context, reservationId uuid.UUID) (*apimodel.ReservationResponse, error) {
	return r.getByIDResp, r.getByIDErr
}
func (r *handlerReservationOrchestratorStub) GetReservationByQuoteId(ctx context.Context, quoteId string) (*apimodel.ReservationResponse, error) {
	return r.getByQuoteResp, r.getByQuoteErr
}
func (r *handlerReservationOrchestratorStub) GetReservationByOrderId(ctx context.Context, orderId string) (*apimodel.ReservationResponse, error) {
	panic("not used")
}
func (r *handlerReservationOrchestratorStub) CommitReservation(ctx context.Context, params apimodel.CommitReservationRequest) (*model.Reservation, error) {
	panic("not used")
}
func (r *handlerReservationOrchestratorStub) ReleaseReservation(ctx context.Context, id uuid.UUID) error {
	panic("not used")
}
func (r *handlerReservationOrchestratorStub) RevertReservation(ctx context.Context, request apimodel.RevertReservationRequest) error {
	panic("not used")
}
func (r *handlerReservationOrchestratorStub) ProcessExpiredReservations(ctx context.Context) (successCounter int, failureCounter int, err error) {
	panic("not used")
}

type handlerReservationServiceStub struct{}

func (s *handlerReservationServiceStub) GetReservationById(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) GetReservationByIdForUpdate(ctx context.Context, id uuid.UUID, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) GetReservationByQuoteId(ctx context.Context, quoteId string) (*model.Reservation, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) GetReservationByOrderId(ctx context.Context, orderId string) (*model.Reservation, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) GetToBeExpiredReservations(ctx context.Context) ([]*model.Reservation, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) AttachOrder(ctx context.Context, request apimodel.AttachOrderRequest) error {
	panic("not used")
}
func (s *handlerReservationServiceStub) ArchiveReservations(ctx context.Context) (int, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) CreateReservationHelper(ctx context.Context, request apimodel.CreateReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) UpdateReservationHelper(ctx context.Context, request apimodel.UpdateReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) CommitReservationHelper(ctx context.Context, request apimodel.CommitReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) ReleaseReservationHelper(ctx context.Context, id uuid.UUID, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) RevertReservationHelper(ctx context.Context, request apimodel.RevertReservationRequest, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (s *handlerReservationServiceStub) ExpireReservationHelper(ctx context.Context, reservation *model.Reservation, u *uow.UnitOfWork) error {
	panic("not used")
}

func decodeResp(t *testing.T, rec *httptest.ResponseRecorder) apimodel.APIResponse {
	t.Helper()
	var out apimodel.APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return out
}

func TestStockHandlerRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New("error", "text")

	engine := gin.New()
	stockSvc := &handlerStockServiceStub{
		getBySkuResp: &apimodel.StockResponse{SKU: "SKU-1", OnHand: 5, Reserved: 1, Available: 4},
		getStocksResp: []*apimodel.StockResponse{
			{SKU: "SKU-1", OnHand: 5, Reserved: 1, Available: 4},
		},
		getStocksPagination: &apimodel.PaginationResponse{Limit: 50, Offset: 0, TotalItems: 1, CurrentPage: 1, TotalPages: 1},
		getStocksMessage:    "Page 1 of 1",
	}
	sh := NewStockHandler(stockSvc, log)
	engine.GET("/stock/:sku", sh.GetStockBySku)
	engine.GET("/stock", sh.GetStocks)

	req := httptest.NewRequest(http.MethodGet, "/stock/SKU-1", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/stock?limit=10&offset=1", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	stockSvc.getBySkuResp = nil
	req = httptest.NewRequest(http.MethodGet, "/stock/SKU-404", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	stockSvc.getBySkuErr = apperror.New(apperror.CodeDBNoRowsCode, apperror.CodeDBNoRowsMessage)
	req = httptest.NewRequest(http.MethodGet, "/stock/SKU-ERR", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for app error, got %d", rec.Code)
	}
}

func TestAdminHandlerRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New("error", "text")
	engine := gin.New()

	adminOrch := &handlerAdminOrchestratorStub{
		adjustResp: &model.Stock{SKU: "SKU-1", OnHand: 10, Reserved: 2},
		activeItems: []*model.ReservationItem{
			{ReservationId: uuid.New(), SKU: "SKU-1", Qty: 1, IsActive: true},
		},
		activePage: &apimodel.PaginationResponse{Limit: 50, Offset: 0, TotalItems: 1, CurrentPage: 1, TotalPages: 1},
		activeMsg:  "Page 1 of 1",
	}
	stockSvc := &handlerStockServiceStub{
		createResp: &model.Stock{SKU: "SKU-1", OnHand: 10, Reserved: 2},
	}
	ah := NewAdminHandler(adminOrch, stockSvc, log)

	engine.POST("/admin/stock", ah.CreateStock)
	engine.PUT("/admin/stock", ah.UpdateStock)
	engine.DELETE("/admin/stock/:sku", ah.DeleteStock)
	engine.GET("/admin/stock/:sku/reservation-items", ah.GetActiveReservationItemsBySku)

	body := bytes.NewBufferString(`{"sku":"SKU-1","on_hand":10,"reserved":2}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/stock", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body = bytes.NewBufferString(`{"sku":"SKU-1","on_hand":11}`)
	req = httptest.NewRequest(http.MethodPut, "/admin/stock", body)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/admin/stock/SKU-1", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/admin/stock/SKU-1/reservation-items", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	adminOrch.deleteErr = errors.New("cannot delete")
	req = httptest.NewRequest(http.MethodDelete, "/admin/stock/SKU-1", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestReservationHandlerValidationRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New("error", "text")
	engine := gin.New()
	resID := uuid.New()
	ro := &handlerReservationOrchestratorStub{
		getByIDResp: &apimodel.ReservationResponse{
			Reservation: &model.Reservation{ReservationId: resID, QuoteId: "Q-1"},
			Items:       map[string]*model.ReservationItem{},
		},
		getByQuoteResp: &apimodel.ReservationResponse{
			Reservation: &model.Reservation{ReservationId: resID, QuoteId: "Q-1"},
			Items:       map[string]*model.ReservationItem{},
		},
	}
	rh := NewReservationHandler(ro, &handlerReservationServiceStub{}, log)

	engine.GET("/reservation/:id", rh.GetReservationById)
	engine.GET("/reservation/quote/:quote_id", rh.GetReservationByQuoteId)

	req := httptest.NewRequest(http.MethodGet, "/reservation/"+resID.String(), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/reservation/not-a-uuid", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	resp := decodeResp(t, rec)
	if resp.Error == nil || resp.Error.Code != apperror.CodeValidationErrorCode {
		t.Fatalf("unexpected error response: %+v", resp.Error)
	}

	req = httptest.NewRequest(http.MethodGet, "/reservation/quote/Q-1", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ro.getByQuoteErr = errors.New("not found")
	req = httptest.NewRequest(http.MethodGet, "/reservation/quote/Q-404", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandlersPoolConstructor(t *testing.T) {
	log := logger.New("error", "text")
	pool := NewHandlersPool(
		&handlerReservationOrchestratorStub{},
		&handlerAdminOrchestratorStub{},
		&handlerStockServiceStub{},
		&handlerReservationServiceStub{},
		log,
	)
	if pool == nil || pool.Stock == nil || pool.Reservation == nil || pool.Admin == nil {
		t.Fatalf("expected full handlers pool")
	}

	// Compile-time assertion for interfaces used by constructors.
	var _ application.ReservationOrchestratorInterface = &handlerReservationOrchestratorStub{}
	var _ application.AdminStockOrchestratorInterface = &handlerAdminOrchestratorStub{}
}
