package repository

import (
	"context"
	"go-inventory-reservations/internal/model"
)

// StockRepository defines operations for stock management
type StockRepository interface {
	GetByID(ctx context.Context, sku string) (*model.Stock, error)
	List(ctx context.Context, limit, offset int) ([]*model.Stock, error)
	Save(ctx context.Context, stock *model.Stock) error
	Update(ctx context.Context, stock *model.Stock) error
	UpdateQuantity(ctx context.Context, sku string, onHand, reserved int) error
}
