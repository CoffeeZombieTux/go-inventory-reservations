package service

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/apperror"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/repository"
	"go-inventory-reservations/internal/uow"
)

// StockServiceInterface defines business operations for stock management
type StockServiceInterface interface {
	CreateStock(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error)
	GetStockBySku(ctx context.Context, sku string) (*apimodel.StockResponse, error)
	GetStockBySkuForUpdate(ctx context.Context, sku string, uow *uow.UnitOfWork) (*model.Stock, error)
	GetStocks(
		ctx context.Context,
		requestedLimit,
		requestedOffset int,
	) (
		stocksResp []*apimodel.StockResponse,
		pagination *apimodel.PaginationResponse,
		message string,
		err error,
	)
	ReserveStock(ctx context.Context, sku string, qty int, uow *uow.UnitOfWork) (*model.Stock, error)
	AdjustInventory(ctx context.Context, req apimodel.StockRequest, uow *uow.UnitOfWork) (*model.Stock, error)
	DeleteStock(ctx context.Context, sku string) error
	CalculateAvailability(ctx context.Context, stock *model.Stock) int
}

// StockService is a service for stock management.
type StockService struct {
	repo repository.StockRepositoryInterface
}

// NewStockService creates a new StockService instance.
func NewStockService(repo repository.StockRepositoryInterface) StockServiceInterface {
	return &StockService{
		repo: repo,
	}
}

// CreateStock creates a new stock item.
func (ss *StockService) CreateStock(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error) {
	var (
		onHand   int
		reserved int
	)
	if req.OnHand != nil {
		onHand = *req.OnHand
	}
	if req.Reserved != nil {
		reserved = *req.Reserved
	}
	stock := &model.Stock{SKU: req.SKU, OnHand: onHand, Reserved: reserved}
	res, err := ss.repo.Create(ctx, stock)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetStockBySku returns a stock item by its SKU.
func (ss *StockService) GetStockBySku(ctx context.Context, sku string) (*apimodel.StockResponse, error) {
	stock, err := ss.repo.GetBySku(ctx, sku)
	if err != nil {
		return nil, err
	}
	resp := apimodel.StockResponse{
		SKU:       stock.SKU,
		OnHand:    stock.OnHand,
		Reserved:  stock.Reserved,
		Available: ss.CalculateAvailability(ctx, stock),
	}
	return &resp, nil
}

// GetStockBySkuForUpdate returns a stock item by its SKU with a FOR UPDATE lock.
func (ss *StockService) GetStockBySkuForUpdate(
	ctx context.Context,
	sku string,
	uow *uow.UnitOfWork,
) (*model.Stock, error) {
	return ss.repo.GetBySkuForUpdate(ctx, sku, uow)
}

// GetStocks returns a list of stock items.
func (ss *StockService) GetStocks(
	ctx context.Context,
	requestedLimit,
	requestedOffset int,
) (
	stocksResp []*apimodel.StockResponse,
	pagination *apimodel.PaginationResponse,
	message string,
	err error,
) {
	// Apply business rules to pagination parameters
	params := apimodel.NewPaginationParams(requestedLimit, requestedOffset)
	stocks, err := ss.repo.GetStocks(ctx, params.Limit, params.Offset)
	if err != nil {
		return nil, nil, "", err
	}

	totalCount, err := ss.repo.Count(ctx)
	if err != nil {
		return nil, nil, "", err
	}

	if totalCount == 0 {
		message = "No stocks found"
		return []*apimodel.StockResponse{}, nil, message, nil
	}

	for _, stock := range stocks {
		stocksResp = append(stocksResp, &apimodel.StockResponse{
			SKU:       stock.SKU,
			OnHand:    stock.OnHand,
			Reserved:  stock.Reserved,
			Available: ss.CalculateAvailability(ctx, stock),
		})
	}

	currentPage := params.Offset/params.Limit + 1
	totalPages := (totalCount + params.Limit - 1) / params.Limit

	if currentPage > totalPages {
		message = fmt.Sprintf("Page %d does not exist. Total pages: %d", currentPage, totalPages)
		return []*apimodel.StockResponse{}, nil, message, nil
	}

	pagination = &apimodel.PaginationResponse{
		Limit:       params.Limit,
		Offset:      params.Offset,
		TotalItems:  totalCount,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}

	message = fmt.Sprintf("Page %d of %d", currentPage, totalPages)

	return stocksResp, pagination, message, nil
}

// AdjustInventory updates the stock quantity for a given SKU.
func (ss *StockService) AdjustInventory(
	ctx context.Context,
	req apimodel.StockRequest,
	uow *uow.UnitOfWork,
) (*model.Stock, error) {
	return ss.repo.Update(ctx, req, uow)
}

// ReserveStock reserves a given quantity of stock items for a given SKU. Qty can be negative.
func (ss *StockService) ReserveStock(
	ctx context.Context,
	sku string,
	qty int,
	uow *uow.UnitOfWork,
) (*model.Stock, error) {
	stock, err := ss.GetStockBySkuForUpdate(ctx, sku, uow)
	if err != nil {
		return nil, err
	}
	newReserved := stock.Reserved + qty
	if newReserved < 0 {
		return nil, apperror.New(
			apperror.CodeValidationError,
			"Cannot reserve negative quantity",
			"cannot reserve negative quantity for SKU "+stock.SKU,
		)
	}
	req := apimodel.StockRequest{
		SKU:      sku,
		Reserved: &newReserved,
	}
	return ss.repo.Update(ctx, req, uow)
}

// DeleteStock deletes a stock item by SKU.
func (ss *StockService) DeleteStock(ctx context.Context, sku string) error {
	return ss.repo.Delete(ctx, sku)
}

// CalculateAvailability calculates the available quantity for a given stock item.
func (ss *StockService) CalculateAvailability(ctx context.Context, stock *model.Stock) int {
	return stock.OnHand - stock.Reserved
}
