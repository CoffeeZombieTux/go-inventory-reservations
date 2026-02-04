package service

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/repository"
)

// StockServiceInterface defines business operations for stock management
type StockServiceInterface interface {
	CreateStock(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error)
	GetStockBySku(ctx context.Context, sku string) (*model.Stock, error)
	GetStocks(
		ctx context.Context,
		requestedLimit,
		requestedOffset int,
	) (
		stocks []*model.Stock,
		pagination *apimodel.PaginationResponse,
		message string,
		err error,
	)
	AdjustInventory(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error)
	DeleteStock(ctx context.Context, sku string) error
}

type StockService struct {
	repo repository.StockRepositoryInterface
}

func NewStockService(repo repository.StockRepositoryInterface) StockServiceInterface {
	return &StockService{
		repo: repo,
	}
}

func (ss *StockService) CreateStock(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error) {
	stock := &model.Stock{SKU: req.SKU, OnHand: req.Quantity}
	res, err := ss.repo.Save(ctx, stock)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ss *StockService) GetStockBySku(ctx context.Context, sku string) (*model.Stock, error) {
	return ss.repo.GetBySku(ctx, sku)
}

func (ss *StockService) GetStocks(
	ctx context.Context,
	requestedLimit,
	requestedOffset int,
) (
	stocks []*model.Stock,
	pagination *apimodel.PaginationResponse,
	message string,
	err error,
) {
	// Apply business rules to pagination parameters
	params := apimodel.NewPaginationParams(requestedLimit, requestedOffset)
	stocks, err = ss.repo.GetStocks(ctx, params.Limit, params.Offset)
	if err != nil {
		return nil, nil, "", err
	}

	totalCount, err := ss.repo.Count(ctx)
	if err != nil {
		return nil, nil, "", err
	}

	if totalCount == 0 {
		message = "No stocks found"
		return []*model.Stock{}, nil, message, nil
	}

	currentPage := params.Offset/params.Limit + 1
	totalPages := (totalCount + params.Limit - 1) / params.Limit

	if currentPage > totalPages {
		message = fmt.Sprintf("Page %d does not exist. Total pages: %d", currentPage, totalPages)
		return []*model.Stock{}, nil, message, nil
	}

	pagination = &apimodel.PaginationResponse{
		Limit:       params.Limit,
		Offset:      params.Offset,
		TotalItems:  totalCount,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}

	message = fmt.Sprintf("Page %d of %d", currentPage, totalPages)

	return stocks, pagination, message, nil
}

func (ss *StockService) AdjustInventory(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error) {
	return ss.repo.UpdateQuantity(ctx, req.SKU, req.Quantity)
}

func (ss *StockService) DeleteStock(ctx context.Context, sku string) error {
	return ss.repo.Delete(ctx, sku)
}
