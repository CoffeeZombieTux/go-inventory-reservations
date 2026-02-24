package application

import (
	"context"
	"errors"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"
	"testing"

	"github.com/google/uuid"
)

type appStockServiceStub struct {
	getBySkuForUpdateStock *model.Stock
	getBySkuForUpdateErr   error
	calculateAvailability  int

	reserveStockCalls int
	lastReserveSKU    string
	lastReserveQty    int
	reserveStockResp  *model.Stock
	reserveStockErr   error

	adjustCalls int
	lastAdjust  apimodel.StockRequest
	adjustResp  *model.Stock
	adjustErr   error

	deleteCalls int
}

func (s *appStockServiceStub) CreateStock(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error) {
	panic("not used")
}
func (s *appStockServiceStub) GetStockBySku(ctx context.Context, sku string) (*apimodel.StockResponse, error) {
	panic("not used")
}
func (s *appStockServiceStub) GetStockBySkuForUpdate(ctx context.Context, sku string, u *uow.UnitOfWork) (*model.Stock, error) {
	if s.getBySkuForUpdateErr != nil {
		return nil, s.getBySkuForUpdateErr
	}
	return s.getBySkuForUpdateStock, nil
}
func (s *appStockServiceStub) GetStocks(
	ctx context.Context,
	requestedLimit,
	requestedOffset int,
) ([]*apimodel.StockResponse, *apimodel.PaginationResponse, string, error) {
	panic("not used")
}
func (s *appStockServiceStub) ReserveStock(ctx context.Context, sku string, qty int, u *uow.UnitOfWork) (*model.Stock, error) {
	s.reserveStockCalls++
	s.lastReserveSKU = sku
	s.lastReserveQty = qty
	if s.reserveStockErr != nil {
		return nil, s.reserveStockErr
	}
	if s.reserveStockResp != nil {
		return s.reserveStockResp, nil
	}
	return &model.Stock{SKU: sku, Reserved: qty}, nil
}
func (s *appStockServiceStub) AdjustInventory(ctx context.Context, req apimodel.StockRequest, u *uow.UnitOfWork) (*model.Stock, error) {
	s.adjustCalls++
	s.lastAdjust = req
	if s.adjustErr != nil {
		return nil, s.adjustErr
	}
	if s.adjustResp != nil {
		return s.adjustResp, nil
	}
	return &model.Stock{SKU: req.SKU}, nil
}
func (s *appStockServiceStub) DeleteStock(ctx context.Context, sku string) error {
	s.deleteCalls++
	return nil
}
func (s *appStockServiceStub) CalculateAvailability(ctx context.Context, stock *model.Stock) int {
	if s.calculateAvailability != 0 {
		return s.calculateAvailability
	}
	return stock.OnHand - stock.Reserved
}

type appReservationServiceStub struct {
	getByIDResp    *model.Reservation
	getByIDErr     error
	getByQuoteResp *model.Reservation
	getByQuoteErr  error
	getByOrderResp *model.Reservation
	getByOrderErr  error
}

func (s *appReservationServiceStub) GetReservationById(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	return s.getByIDResp, s.getByIDErr
}
func (s *appReservationServiceStub) GetReservationByIdForUpdate(ctx context.Context, id uuid.UUID, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (s *appReservationServiceStub) GetReservationByQuoteId(ctx context.Context, quoteId string) (*model.Reservation, error) {
	return s.getByQuoteResp, s.getByQuoteErr
}
func (s *appReservationServiceStub) GetReservationByOrderId(ctx context.Context, orderId string) (*model.Reservation, error) {
	return s.getByOrderResp, s.getByOrderErr
}
func (s *appReservationServiceStub) GetToBeExpiredReservations(ctx context.Context) ([]*model.Reservation, error) {
	panic("not used")
}
func (s *appReservationServiceStub) AttachOrder(ctx context.Context, request apimodel.AttachOrderRequest) error {
	panic("not used")
}
func (s *appReservationServiceStub) ArchiveReservations(ctx context.Context) (int, error) {
	panic("not used")
}
func (s *appReservationServiceStub) CreateReservationHelper(
	ctx context.Context,
	request apimodel.CreateReservationRequest,
	u *uow.UnitOfWork,
) (*model.Reservation, error) {
	panic("not used")
}
func (s *appReservationServiceStub) UpdateReservationHelper(
	ctx context.Context,
	request apimodel.UpdateReservationRequest,
	u *uow.UnitOfWork,
) (*model.Reservation, error) {
	panic("not used")
}
func (s *appReservationServiceStub) CommitReservationHelper(
	ctx context.Context,
	request apimodel.CommitReservationRequest,
	u *uow.UnitOfWork,
) (*model.Reservation, error) {
	panic("not used")
}
func (s *appReservationServiceStub) ReleaseReservationHelper(ctx context.Context, id uuid.UUID, u *uow.UnitOfWork) (*model.Reservation, error) {
	panic("not used")
}
func (s *appReservationServiceStub) RevertReservationHelper(
	ctx context.Context,
	request apimodel.RevertReservationRequest,
	u *uow.UnitOfWork,
) (*model.Reservation, error) {
	panic("not used")
}
func (s *appReservationServiceStub) ExpireReservationHelper(ctx context.Context, reservation *model.Reservation, u *uow.UnitOfWork) error {
	panic("not used")
}

type appReservationItemsServiceStub struct {
	getItemsResp map[string]*model.ReservationItem
	getItemsErr  error

	createResp *model.ReservationItem
	createErr  error

	updateResp *model.ReservationItem
	updateErr  error

	setActiveCalls int
	setActiveErr   error

	deleteCalls int
	deleteErr   error

	activeBySKU map[string]*model.ReservationItem
	activeErr   error
}

func (s *appReservationItemsServiceStub) GetReservationItem(
	ctx context.Context,
	reservationId uuid.UUID,
	sku string,
	u *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	panic("not used")
}
func (s *appReservationItemsServiceStub) GetReservationItems(
	ctx context.Context,
	reservationId uuid.UUID,
	u *uow.UnitOfWork,
) (map[string]*model.ReservationItem, error) {
	return s.getItemsResp, s.getItemsErr
}
func (s *appReservationItemsServiceStub) GetActiveReservationItemsBySku(
	ctx context.Context,
	sku string,
	u *uow.UnitOfWork,
) (map[string]*model.ReservationItem, error) {
	return s.activeBySKU, s.activeErr
}
func (s *appReservationItemsServiceStub) CreateReservationItem(
	ctx context.Context,
	request apimodel.ReservationItemRequest,
	reservationId uuid.UUID,
	u *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	if s.createResp != nil {
		return s.createResp, nil
	}
	return &model.ReservationItem{ReservationId: reservationId, SKU: request.SKU, Qty: request.Quantity, IsActive: true}, nil
}
func (s *appReservationItemsServiceStub) UpdateReservationItem(
	ctx context.Context,
	request apimodel.ReservationItemRequest,
	reservationId uuid.UUID,
	u *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	if s.updateResp != nil {
		return s.updateResp, nil
	}
	return &model.ReservationItem{ReservationId: reservationId, SKU: request.SKU, Qty: request.Quantity, IsActive: request.Quantity > 0}, nil
}
func (s *appReservationItemsServiceStub) SetReservationItemActive(
	ctx context.Context,
	reservationId uuid.UUID,
	sku string,
	isActive bool,
	u *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	s.setActiveCalls++
	if s.setActiveErr != nil {
		return nil, s.setActiveErr
	}
	return &model.ReservationItem{ReservationId: reservationId, SKU: sku, IsActive: isActive}, nil
}
func (s *appReservationItemsServiceStub) DeleteReservationItem(
	ctx context.Context,
	reservationId uuid.UUID,
	sku string,
	u *uow.UnitOfWork,
) error {
	s.deleteCalls++
	return s.deleteErr
}

func TestGetReservationMethods(t *testing.T) {
	id := uuid.New()
	res := &model.Reservation{ReservationId: id}
	itemsSvc := &appReservationItemsServiceStub{
		getItemsResp: map[string]*model.ReservationItem{
			"SKU-1": {ReservationId: id, SKU: "SKU-1", Qty: 1, IsActive: true},
		},
	}
	ro := &ReservationOrchestrator{
		reservationService:     &appReservationServiceStub{getByIDResp: res, getByQuoteResp: res, getByOrderResp: res},
		reservationItemService: itemsSvc,
	}

	if _, err := ro.GetReservationById(context.Background(), id); err != nil {
		t.Fatalf("unexpected GetReservationById err: %v", err)
	}
	if _, err := ro.GetReservationByQuoteId(context.Background(), "Q-1"); err != nil {
		t.Fatalf("unexpected GetReservationByQuoteId err: %v", err)
	}
	if _, err := ro.GetReservationByOrderId(context.Background(), "O-1"); err != nil {
		t.Fatalf("unexpected GetReservationByOrderId err: %v", err)
	}

	ro.reservationItemService = &appReservationItemsServiceStub{getItemsErr: errors.New("items failed")}
	if _, err := ro.GetReservationByOrderId(context.Background(), "O-1"); err == nil {
		t.Fatalf("expected GetReservationByOrderId to propagate items error")
	}
}

func TestUpdateReservationHelpers(t *testing.T) {
	reservationID := uuid.New()
	stockSvc := &appStockServiceStub{
		getBySkuForUpdateStock: &model.Stock{SKU: "SKU-1", OnHand: 10, Reserved: 2},
	}
	itemsSvc := &appReservationItemsServiceStub{}
	ro := &ReservationOrchestrator{
		stockService:           stockSvc,
		reservationItemService: itemsSvc,
	}

	resultItems := map[string]*model.ReservationItem{}
	existing := &model.ReservationItem{ReservationId: reservationID, SKU: "SKU-1", Qty: 1, IsActive: true}
	if err := ro.applyExistingReservationItemUpdate(
		context.Background(),
		reservationID,
		apimodel.ReservationItemRequest{SKU: "SKU-1", Quantity: 1},
		existing,
		stockSvc.getBySkuForUpdateStock,
		resultItems,
		nil,
	); err != nil {
		t.Fatalf("unexpected no-op update error: %v", err)
	}
	if resultItems["SKU-1"] == nil {
		t.Fatalf("expected existing item to be reused")
	}

	stockSvc.reserveStockResp = &model.Stock{SKU: "SKU-1", Reserved: 0}
	if err := ro.applyExistingReservationItemUpdate(
		context.Background(),
		reservationID,
		apimodel.ReservationItemRequest{SKU: "SKU-1", Quantity: 0},
		existing,
		stockSvc.getBySkuForUpdateStock,
		map[string]*model.ReservationItem{},
		nil,
	); err != nil {
		t.Fatalf("unexpected update-delete error: %v", err)
	}
	if itemsSvc.deleteCalls == 0 {
		t.Fatalf("expected delete call for zero reserved stock")
	}

	stockSvc.calculateAvailability = 0
	_, err := ro.adjustReservedStockForUpdate(context.Background(), "SKU-1", 2, &model.Stock{SKU: "SKU-1", OnHand: 1, Reserved: 1}, nil)
	if err == nil {
		t.Fatalf("expected insufficient stock error")
	}
}

func TestApplyNewAndRemoveMissingReservationItems(t *testing.T) {
	reservationID := uuid.New()
	stockSvc := &appStockServiceStub{
		getBySkuForUpdateStock: &model.Stock{SKU: "SKU-2", OnHand: 10, Reserved: 3},
	}
	itemsSvc := &appReservationItemsServiceStub{}
	ro := &ReservationOrchestrator{
		stockService:           stockSvc,
		reservationItemService: itemsSvc,
	}

	result := map[string]*model.ReservationItem{}
	if err := ro.applyNewReservationItem(context.Background(), reservationID, apimodel.ReservationItemRequest{
		SKU:      "SKU-2",
		Quantity: 2,
	}, result, nil); err != nil {
		t.Fatalf("unexpected applyNewReservationItem error: %v", err)
	}
	if result["SKU-2"] == nil || stockSvc.reserveStockCalls == 0 {
		t.Fatalf("expected new item and reserve stock call")
	}

	missing := map[string]*model.ReservationItem{
		"SKU-X": {ReservationId: reservationID, SKU: "SKU-X", Qty: 3, IsActive: true},
	}
	if err := ro.removeMissingReservationItems(context.Background(), missing, nil); err != nil {
		t.Fatalf("unexpected removeMissingReservationItems error: %v", err)
	}
	if itemsSvc.deleteCalls == 0 {
		t.Fatalf("expected delete call for missing item")
	}
}

func TestCommitAndReleaseHelpers(t *testing.T) {
	reservationID := uuid.New()
	stockSvc := &appStockServiceStub{
		getBySkuForUpdateStock: &model.Stock{SKU: "SKU-3", OnHand: 10, Reserved: 5},
	}
	itemsSvc := &appReservationItemsServiceStub{
		getItemsResp: map[string]*model.ReservationItem{
			"SKU-3": {ReservationId: reservationID, SKU: "SKU-3", Qty: 2, IsActive: true},
		},
	}
	ro := &ReservationOrchestrator{
		stockService:           stockSvc,
		reservationItemService: itemsSvc,
	}

	if _, err := ro.getReservationItemsForCommit(context.Background(), reservationID, nil); err != nil {
		t.Fatalf("unexpected getReservationItemsForCommit error: %v", err)
	}
	itemsSvc.getItemsResp = map[string]*model.ReservationItem{}
	if _, err := ro.getReservationItemsForCommit(context.Background(), reservationID, nil); err == nil {
		t.Fatalf("expected empty-items commit error")
	}

	item := &model.ReservationItem{ReservationId: reservationID, SKU: "SKU-3", Qty: 2, IsActive: true}
	if err := ro.commitReservationItem(context.Background(), reservationID, item, nil); err != nil {
		t.Fatalf("unexpected commitReservationItem error: %v", err)
	}
	if stockSvc.adjustCalls == 0 || itemsSvc.setActiveCalls == 0 {
		t.Fatalf("expected stock adjust and set inactive calls")
	}

	itemsSvc.getItemsResp = map[string]*model.ReservationItem{"SKU-3": item}
	if err := ro.releaseReservationStocks(context.Background(), reservationID, nil); err != nil {
		t.Fatalf("unexpected releaseReservationStocks error: %v", err)
	}
}

func TestAdminStockOrchestratorNonTransactionalMethods(t *testing.T) {
	stockSvc := &appStockServiceStub{}
	itemsSvc := &appReservationItemsServiceStub{
		activeBySKU: map[string]*model.ReservationItem{
			uuid.New().String(): {SKU: "SKU-1", Qty: 1, IsActive: true},
		},
	}
	a := AdminStockOrchestrator{
		stockService:           stockSvc,
		reservationItemService: itemsSvc,
	}

	if err := a.DeleteStock(context.Background(), "SKU-1"); err == nil {
		t.Fatalf("expected delete to fail when active reservations exist")
	}

	itemsSvc.activeBySKU = map[string]*model.ReservationItem{}
	if err := a.DeleteStock(context.Background(), "SKU-1"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if stockSvc.deleteCalls == 0 {
		t.Fatalf("expected stock delete call")
	}

	itemsSvc.activeBySKU = map[string]*model.ReservationItem{
		uuid.MustParse("11111111-1111-1111-1111-111111111111").String(): {
			ReservationId: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			SKU:           "SKU-1",
			Qty:           1,
			IsActive:      true,
		},
		uuid.MustParse("22222222-2222-2222-2222-222222222222").String(): {
			ReservationId: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			SKU:           "SKU-1",
			Qty:           2,
			IsActive:      true,
		},
	}

	items, pagination, message, err := a.GetActiveReservationItemsBySku(context.Background(), "SKU-1", 1, 0)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if len(items) != 1 || pagination == nil || message == "" {
		t.Fatalf("unexpected pagination result: items=%d pagination=%+v message=%q", len(items), pagination, message)
	}

	items, pagination, message, err = a.GetActiveReservationItemsBySku(context.Background(), "SKU-1", 1, 10)
	if err != nil {
		t.Fatalf("unexpected out-of-range list error: %v", err)
	}
	if len(items) != 0 || pagination != nil || message == "" {
		t.Fatalf("unexpected out-of-range result: items=%d pagination=%+v message=%q", len(items), pagination, message)
	}
}

func TestReservationOrchestratorConstructor(t *testing.T) {
	ro := NewReservationOrchestrator(
		nil,
		&appStockServiceStub{},
		&appReservationServiceStub{},
		&appReservationItemsServiceStub{},
		nil,
		nil,
	)
	if ro == nil {
		t.Fatalf("expected orchestrator")
	}
}

func TestAdminStockOrchestratorConstructor(t *testing.T) {
	cfg := &config.Config{}
	_ = cfg
	o := NewAdminStockOrchestrator(nil, &appStockServiceStub{}, &appReservationItemsServiceStub{})
	if o == nil {
		t.Fatalf("expected orchestrator")
	}
}
