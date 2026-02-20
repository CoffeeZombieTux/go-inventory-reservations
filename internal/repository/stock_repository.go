package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"
	"strings"
	"time"
)

// StockRepositoryInterface defines operations for stock management
type StockRepositoryInterface interface {
	GetBySku(ctx context.Context, sku string, uow *uow.UnitOfWork) (*model.Stock, error)
	GetBySkuForUpdate(ctx context.Context, sku string, uow *uow.UnitOfWork) (*model.Stock, error)
	GetStocks(ctx context.Context, limit, offset int) ([]*model.Stock, error)
	Create(ctx context.Context, stock *model.Stock) (*model.Stock, error)
	Update(ctx context.Context, request apimodel.StockRequest, uow *uow.UnitOfWork) (*model.Stock, error)
	Delete(ctx context.Context, sku string) error
	Count(ctx context.Context) (int, error)
}

// StockRepository is a repository for stock management.
type StockRepository struct {
	db *sql.DB
}

// NewStockRepository creates a new StockRepository instance.
func NewStockRepository(db *sql.DB) StockRepositoryInterface {
	return &StockRepository{
		db: db,
	}
}

// GetBySku returns a stock by its SKU.
func (sr *StockRepository) GetBySku(ctx context.Context, sku string, uow *uow.UnitOfWork) (*model.Stock, error) {
	return sr.getBySku(ctx, sku, uow, false)
}

// GetBySkuForUpdate returns a stock by its SKU with FOR UPDATE lock.
func (sr *StockRepository) GetBySkuForUpdate(ctx context.Context, sku string, uow *uow.UnitOfWork) (*model.Stock, error) {
	return sr.getBySku(ctx, sku, uow, true)
}

func (sr *StockRepository) getBySku(ctx context.Context, sku string, uow *uow.UnitOfWork, forUpdate bool) (*model.Stock, error) {
	var exec SQLExecutor
	if uow != nil && uow.GetTransaction() != nil {
		exec = uow.GetTransaction()
	} else {
		exec = sr.db
	}

	query := `
        SELECT sku, on_hand, reserved, updated_at
        FROM stock
        WHERE sku = $1
    `
	if forUpdate {
		query += " FOR UPDATE"
	}

	var stock model.Stock
	err := exec.QueryRowContext(
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("stock not found for SKU: %s", sku)
		}
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	return &stock, nil
}

// GetStocks returns a list of stocks.
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

// Save inserts a stock into the database or updates it if it already exists.
func (sr *StockRepository) Create(ctx context.Context, stock *model.Stock) (*model.Stock, error) {
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

// UpdateQuantity updates the stock quantity for a given SKU.
func (sr *StockRepository) Update(ctx context.Context, request apimodel.StockRequest, uow *uow.UnitOfWork) (*model.Stock, error) {
	var exec SQLExecutor
	if uow != nil && uow.GetTransaction() != nil {
		exec = uow.GetTransaction()
	} else {
		exec = sr.db
	}

	var (
		fields []string
		args   []interface{}
		i      = 1
	)

	if request.OnHand != nil {
		fields = append(fields, fmt.Sprintf("on_hand = $%d", i))
		args = append(args, *request.OnHand)
		i++
	}
	if request.Reserved != nil {
		fields = append(fields, fmt.Sprintf("reserved = $%d", i))
		args = append(args, *request.Reserved)
		i++
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("no updatable fields provided")
	}
	args = append(args, request.SKU)

	query := fmt.Sprintf(`
		UPDATE stock SET %s, updated_at = NOW()
		WHERE sku = $%d
		RETURNING sku, on_hand, reserved, updated_at
	`,
		strings.Join(fields, ", "),
		i,
	)

	var stock model.Stock
	err := exec.QueryRowContext(ctx, query, args...).Scan(
		&stock.SKU,
		&stock.OnHand,
		&stock.Reserved,
		&stock.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("stock not found for SKU: %s", request.SKU)
		}
		return nil, fmt.Errorf("failed to update stock quantity: %w", err)
	}

	return &stock, nil
}

// Delete deletes a stock by its SKU.
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

// Count returns the number of stocks records in the database.
func (sr *StockRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM stock"

	err := sr.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count stocks records: %w", err)
	}

	return count, nil
}
