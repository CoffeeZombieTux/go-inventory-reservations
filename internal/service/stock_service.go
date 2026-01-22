package service

import (
	"context"
	"go-inventory-reservations/internal/model"
)

// StockService defines business operations for stock management
type StockService interface {
	GetStock(ctx context.Context, sku string) (*model.Stock, error)
	ListStock(ctx context.Context, page, pageSize int) ([]*model.Stock, error)
	AdjustInventory(ctx context.Context, sku string, quantity int) error
}
