package service

import (
	"context"
	"errors"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"
	"testing"
	"time"
)

type stockRepoStub struct {
	createCalls int
	updateCalls int
	deleteCalls int

	getBySkuForUpdateStock *model.Stock
	getBySkuForUpdateErr   error

	getStocksResp  []*model.Stock
	getStocksErr   error
	getStocksLimit int
	getStocksOff   int

	countResp int
	countErr  error

	lastUpdateReq apimodel.StockRequest
}

func (s *stockRepoStub) GetBySku(ctx context.Context, sku string) (*model.Stock, error) {
	panic("unexpected call")
}

func (s *stockRepoStub) GetBySkuForUpdate(ctx context.Context, sku string, u *uow.UnitOfWork) (*model.Stock, error) {
	if s.getBySkuForUpdateErr != nil {
		return nil, s.getBySkuForUpdateErr
	}
	if s.getBySkuForUpdateStock == nil {
		panic("unexpected call")
	}
	return s.getBySkuForUpdateStock, nil
}

func (s *stockRepoStub) GetStocks(ctx context.Context, limit, offset int) ([]*model.Stock, error) {
	s.getStocksLimit = limit
	s.getStocksOff = offset
	if s.getStocksErr != nil {
		return nil, s.getStocksErr
	}
	return s.getStocksResp, nil
}

func (s *stockRepoStub) Create(ctx context.Context, stock *model.Stock) (*model.Stock, error) {
	s.createCalls++
	stock.UpdatedAt = time.Now()
	return stock, nil
}

func (s *stockRepoStub) Update(ctx context.Context, request apimodel.StockRequest, u *uow.UnitOfWork) (*model.Stock, error) {
	s.updateCalls++
	s.lastUpdateReq = request
	onHand := 0
	reserved := 0
	if request.OnHand != nil {
		onHand = *request.OnHand
	}
	if request.Reserved != nil {
		reserved = *request.Reserved
	}
	return &model.Stock{
		SKU:      request.SKU,
		OnHand:   onHand,
		Reserved: reserved,
	}, nil
}

func (s *stockRepoStub) Delete(ctx context.Context, sku string) error {
	s.deleteCalls++
	return nil
}

func (s *stockRepoStub) Count(ctx context.Context) (int, error) {
	if s.countErr != nil {
		return 0, s.countErr
	}
	return s.countResp, nil
}

func TestCreateStock_RejectsReservedGreaterThanOnHand(t *testing.T) {
	repo := &stockRepoStub{}
	svc := NewStockService(repo)

	onHand := 1
	reserved := 2
	_, err := svc.CreateStock(context.Background(), apimodel.StockRequest{
		SKU:      "SKU-1",
		OnHand:   &onHand,
		Reserved: &reserved,
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected repo.Create not called, got %d calls", repo.createCalls)
	}
}

func TestCreateStock_AcceptsValidInput(t *testing.T) {
	repo := &stockRepoStub{}
	svc := NewStockService(repo)

	onHand := 10
	reserved := 3
	stock, err := svc.CreateStock(context.Background(), apimodel.StockRequest{
		SKU:      "SKU-2",
		OnHand:   &onHand,
		Reserved: &reserved,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected repo.Create called once, got %d", repo.createCalls)
	}
	if stock.SKU != "SKU-2" || stock.OnHand != 10 || stock.Reserved != 3 {
		t.Fatalf("unexpected stock: %+v", stock)
	}
}

func TestAdjustInventory_RejectsNegativeReserved(t *testing.T) {
	repo := &stockRepoStub{}
	svc := NewStockService(repo)

	reserved := -1
	_, err := svc.AdjustInventory(context.Background(), apimodel.StockRequest{
		SKU:      "SKU-3",
		Reserved: &reserved,
	}, nil)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if repo.updateCalls != 0 {
		t.Fatalf("expected repo.Update not called, got %d calls", repo.updateCalls)
	}
}

func TestAdjustInventory_RejectsNegativeOnHand(t *testing.T) {
	repo := &stockRepoStub{}
	svc := NewStockService(repo)

	onHand := -1
	_, err := svc.AdjustInventory(context.Background(), apimodel.StockRequest{
		SKU:    "SKU-3",
		OnHand: &onHand,
	}, nil)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if repo.updateCalls != 0 {
		t.Fatalf("expected repo.Update not called, got %d calls", repo.updateCalls)
	}
}

func TestGetStocks_EmptyResult(t *testing.T) {
	repo := &stockRepoStub{
		getStocksResp: []*model.Stock{},
		countResp:     0,
	}
	svc := NewStockService(repo)

	stocks, pagination, message, err := svc.GetStocks(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stocks) != 0 {
		t.Fatalf("expected no stocks, got %d", len(stocks))
	}
	if pagination != nil {
		t.Fatalf("expected nil pagination for empty result")
	}
	if message != "No stocks found" {
		t.Fatalf("unexpected message: %q", message)
	}
	if repo.getStocksLimit != 50 || repo.getStocksOff != 0 {
		t.Fatalf("unexpected pagination params passed to repo: limit=%d offset=%d", repo.getStocksLimit, repo.getStocksOff)
	}
}

func TestGetStocks_PageOutOfRange(t *testing.T) {
	repo := &stockRepoStub{
		getStocksResp: []*model.Stock{
			{SKU: "SKU-1", OnHand: 10, Reserved: 2},
		},
		countResp: 1,
	}
	svc := NewStockService(repo)

	stocks, pagination, message, err := svc.GetStocks(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stocks) != 0 {
		t.Fatalf("expected no stocks for out-of-range page, got %d", len(stocks))
	}
	if pagination != nil {
		t.Fatalf("expected nil pagination for out-of-range page")
	}
	if message == "" {
		t.Fatalf("expected out-of-range message")
	}
}

func TestGetStocks_ReturnsAvailabilityAndPagination(t *testing.T) {
	repo := &stockRepoStub{
		getStocksResp: []*model.Stock{
			{SKU: "SKU-1", OnHand: 10, Reserved: 2},
			{SKU: "SKU-2", OnHand: 8, Reserved: 3},
		},
		countResp: 2,
	}
	svc := NewStockService(repo)

	stocks, pagination, message, err := svc.GetStocks(context.Background(), 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stocks) != 2 {
		t.Fatalf("expected 2 stocks, got %d", len(stocks))
	}
	if stocks[0].Available != 8 || stocks[1].Available != 5 {
		t.Fatalf("unexpected availability: %+v", stocks)
	}
	if pagination == nil || pagination.TotalItems != 2 || pagination.TotalPages != 1 || pagination.CurrentPage != 1 {
		t.Fatalf("unexpected pagination: %+v", pagination)
	}
	if message != "Page 1 of 1" {
		t.Fatalf("unexpected message: %q", message)
	}
}

func TestGetStocks_PropagatesRepositoryError(t *testing.T) {
	repo := &stockRepoStub{
		getStocksErr: errors.New("db down"),
	}
	svc := NewStockService(repo)

	_, _, _, err := svc.GetStocks(context.Background(), 10, 0)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestReserveStock_RejectsResultingNegativeReserved(t *testing.T) {
	repo := &stockRepoStub{
		getBySkuForUpdateStock: &model.Stock{SKU: "SKU-1", OnHand: 5, Reserved: 1},
	}
	svc := NewStockService(repo)

	_, err := svc.ReserveStock(context.Background(), "SKU-1", -2, nil)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if repo.updateCalls != 0 {
		t.Fatalf("expected repo.Update not called, got %d calls", repo.updateCalls)
	}
}

func TestReserveStock_UpdatesReservedQuantity(t *testing.T) {
	repo := &stockRepoStub{
		getBySkuForUpdateStock: &model.Stock{SKU: "SKU-1", OnHand: 5, Reserved: 1},
	}
	svc := NewStockService(repo)

	stock, err := svc.ReserveStock(context.Background(), "SKU-1", 2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", repo.updateCalls)
	}
	if repo.lastUpdateReq.Reserved == nil || *repo.lastUpdateReq.Reserved != 3 {
		t.Fatalf("expected reserved=3 in update request, got %+v", repo.lastUpdateReq)
	}
	if stock.Reserved != 3 {
		t.Fatalf("unexpected resulting stock: %+v", stock)
	}
}
