package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-inventory-reservations/internal/model"
	"time"
)

// StockRepositoryInterface defines operations for stock management
type StockRepositoryInterface interface {
	GetBySku(ctx context.Context, sku string) (*model.Stock, error)
	GetStocks(ctx context.Context, limit, offset int) ([]*model.Stock, error)
	Save(ctx context.Context, stock *model.Stock) (*model.Stock, error)
	UpdateQuantity(ctx context.Context, sku string, onHand int) (*model.Stock, error)
	Delete(ctx context.Context, sku string) error
	Count(ctx context.Context) (int, error)
}

type StockRepository struct {
	db *sql.DB
}

func NewStockRepository(db *sql.DB) StockRepositoryInterface {
	return &StockRepository{
		db: db,
	}
}

func (sr *StockRepository) GetBySku(ctx context.Context, sku string) (*model.Stock, error) {
	query := `
        SELECT sku, on_hand, reserved, updated_at
        FROM stock
        WHERE sku = $1
    `

	var stock model.Stock
	err := sr.db.QueryRowContext(
		ctx,
		query,
		sku,
	).Scan(
		&stock.SKU,
		&stock.OnHand,
		&stock.Reserved,
		&stock.UpdatedAt,
	)

	if err != nil {
		// Use errors.Is for wrapped error checking
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("stock not found for SKU: %s", sku)
		}
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	return &stock, nil
}

func (sr *StockRepository) GetStocks(ctx context.Context, limit, offset int) ([]*model.Stock, error) {
	query := `
        SELECT sku, on_hand, reserved, updated_at
        FROM stock
        ORDER BY sku
        LIMIT $1 OFFSET $2
    `

	rows, err := sr.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list stocks: %w", err)
	}
	defer rows.Close()

	var stocks []*model.Stock
	for rows.Next() {
		var stock model.Stock
		if err := rows.Scan(
			&stock.SKU,
			&stock.OnHand,
			&stock.Reserved,
			&stock.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan stock row: %w", err)
		}
		stocks = append(stocks, &stock)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return stocks, nil
}

func (sr *StockRepository) Save(ctx context.Context, stock *model.Stock) (*model.Stock, error) {
	query := `
		INSERT INTO stock (
			sku, on_hand, reserved, 
			updated_at
		) VALUES ($1, $2, $3, NOW())
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := sr.db.QueryRowContext(
		ctx,
		query,
		stock.SKU,
		stock.OnHand,
		stock.Reserved,
	).Scan(&updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to save stock: %w", err)
	}

	stock.UpdatedAt = updatedAt

	return stock, nil
}

func (sr *StockRepository) UpdateQuantity(ctx context.Context, sku string, onHand int) (*model.Stock, error) {
	query := `
        UPDATE stock
        SET on_hand = $2, 
            updated_at = NOW()
        WHERE sku = $1
        RETURNING sku, on_hand, reserved, updated_at
    `

	var stock model.Stock
	err := sr.db.QueryRowContext(ctx, query, sku, onHand).Scan(
		&stock.SKU,
		&stock.OnHand,
		&stock.Reserved,
		&stock.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("stock not found for SKU: %s", sku)
		}
		return nil, fmt.Errorf("failed to update stock quantity: %w", err)
	}

	return &stock, nil
}

func (sr *StockRepository) Delete(ctx context.Context, sku string) error {
	query := `DELETE FROM stock WHERE sku = $1 RETURNING sku`

	var deletedSku string
	err := sr.db.QueryRowContext(ctx, query, sku).Scan(&deletedSku)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("stock not found for SKU: %s", sku)
		}
		return fmt.Errorf("failed to delete stock: %w", err)
	}

	return nil
}

func (sr *StockRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM stock"

	err := sr.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count stocks records: %w", err)
	}

	return count, nil
}
