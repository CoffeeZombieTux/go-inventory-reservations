package application

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/service"
	"go-inventory-reservations/internal/uow"
	"sort"
)

// AdminStockOrchestratorInterface defines admin stock operations.
type AdminStockOrchestratorInterface interface {
	DeleteStock(ctx context.Context, sku string) error
	AdjustInventory(ctx context.Context, req apimodel.StockRequest) (*model.Stock, error)
	GetActiveReservationItemsBySku(
		ctx context.Context,
		sku string,
		requestedLimit int,
		requestedOffset int,
	) (
		items []*model.ReservationItem,
		pagination *apimodel.PaginationResponse,
		message string,
		err error,
	)
}

// AdminStockOrchestrator orchestrates admin stock operations.
type AdminStockOrchestrator struct {
	uowManager             *uow.UnitOfWorkManager
	stockService           service.StockServiceInterface
	reservationItemService service.ReservationItemsServiceInterface
}

// NewAdminStockOrchestrator creates a new AdminStockOrchestrator.
func NewAdminStockOrchestrator(
	uow *uow.UnitOfWorkManager,
	stockService service.StockServiceInterface,
	reservationItemService service.ReservationItemsServiceInterface,
) AdminStockOrchestratorInterface {
	return &AdminStockOrchestrator{
		uowManager:             uow,
		stockService:           stockService,
		reservationItemService: reservationItemService,
	}
}

// DeleteStock deletes stock only when there are no active reservations for the SKU.
func (a AdminStockOrchestrator) DeleteStock(ctx context.Context, sku string) error {
	reservationItems, err := a.reservationItemService.GetActiveReservationItemsBySku(ctx, sku, nil)
	if err != nil {
		return err
	}
	if len(reservationItems) > 0 {
		return fmt.Errorf("cannot delete stock item SKU: %s, it is currently reserved by %d reservations",
			sku,
			len(reservationItems),
		)
	}
	return a.stockService.DeleteStock(ctx, sku)
}

// AdjustInventory adjusts stock in a transaction with reservation safety checks.
func (a AdminStockOrchestrator) AdjustInventory(
	ctx context.Context,
	req apimodel.StockRequest,
) (*model.Stock, error) {
	var result *model.Stock
	err := uow.WithUnitOfWork(ctx, a.uowManager, func(unit *uow.UnitOfWork) error {
		stock, err := a.stockService.GetStockBySkuForUpdate(ctx, req.SKU, unit)
		if err != nil {
			return err
		}

		if req.OnHand != nil && *req.OnHand < stock.Reserved {
			return fmt.Errorf("cannot reduce OnHand below Reserved quantity")
		}

		result, err = a.stockService.AdjustInventory(ctx, req, unit)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetActiveReservationItemsBySku returns paginated active reservation items for a SKU.
func (a AdminStockOrchestrator) GetActiveReservationItemsBySku(
	ctx context.Context,
	sku string,
	requestedLimit int,
	requestedOffset int,
) (
	items []*model.ReservationItem,
	pagination *apimodel.PaginationResponse,
	message string,
	err error,
) {
	itemsMap, err := a.reservationItemService.GetActiveReservationItemsBySku(ctx, sku, nil)
	if err != nil {
		return nil, nil, "", err
	}

	params := apimodel.NewPaginationParams(requestedLimit, requestedOffset)
	items = make([]*model.ReservationItem, 0, len(itemsMap))
	for _, item := range itemsMap {
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ReservationId.String() < items[j].ReservationId.String()
	})

	totalCount := len(items)
	if totalCount == 0 {
		return []*model.ReservationItem{}, nil, "No active reservation items found", nil
	}

	currentPage := params.Offset/params.Limit + 1
	totalPages := (totalCount + params.Limit - 1) / params.Limit
	if currentPage > totalPages {
		return []*model.ReservationItem{}, nil, fmt.Sprintf("Page %d does not exist. Total pages: %d", currentPage, totalPages), nil
	}

	start := params.Offset
	end := start + params.Limit
	if end > totalCount {
		end = totalCount
	}

	pagination = &apimodel.PaginationResponse{
		Limit:       params.Limit,
		Offset:      params.Offset,
		TotalItems:  totalCount,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}

	message = fmt.Sprintf("Page %d of %d", currentPage, totalPages)
	return items[start:end], pagination, message, nil
}
