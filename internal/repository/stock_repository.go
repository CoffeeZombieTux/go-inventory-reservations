package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go-inventory-reservations/internal/apperror"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"
	"strings"
	"time"
)

// StockRepositoryInterface defines operations for stock management
type StockRepositoryInterface interface {
	GetBySku(ctx context.Context, sku string) (*model.Stock, error)
	GetBySkuForUpdate(ctx context.Context, sku string, uow *uow.UnitOfWork) (*model.Stock, error)
	GetStocks(ctx context.Context, limit, offset int) ([]*model.Stock, error)
	Create(ctx context.Context, stock *model.Stock) (*model.Stock, error)
	Update(ctx context.Context, request apimodel.StockRequest, uow *uow.UnitOfWork) (*model.Stock, error)
	Delete(ctx context.Context, sku string) error
	Count(ctx context.Context) (int, error)
}

// StockRepository is a repository for stock management.
type StockRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewStockRepository creates a new StockRepository instance.
func NewStockRepository(db *sql.DB, log *logger.Logger) StockRepositoryInterface {
	return &StockRepository{
		db:     db,
		logger: log,
	}
}

// GetBySku returns a stock by its SKU.
func (sr *StockRepository) GetBySku(ctx context.Context, sku string) (*model.Stock, error) {
	return sr.getBySku(ctx, sku, nil, false)
}

// GetBySkuForUpdate returns a stock by its SKU with FOR UPDATE lock.
func (sr *StockRepository) GetBySkuForUpdate(ctx context.Context, sku string, uow *uow.UnitOfWork) (*model.Stock, error) {
	return sr.getBySku(ctx, sku, uow, true)
}

// getBySku returns a stock record by SKU and optionally applies FOR UPDATE locking.
func (sr *StockRepository) getBySku(
	ctx context.Context,
	sku string,
	uow *uow.UnitOfWork,
	forUpdate bool,
) (*model.Stock, error) {
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
		return nil, apperror.FromDB(err, "Failed to get stock", apperror.CodeInternalError)
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
		return nil, apperror.FromDB(err, "Failed to list stocks", apperror.CodeInternalError)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			sr.logger.WithError(closeErr).WithFields(logger.Fields{
				"repository": "StockRepository",
				"method":     "GetStocks",
			}).Error(logger.LogMessageFailedToCloseRows)
		}
	}()

	var stocks []*model.Stock
	for rows.Next() {
		var stock model.Stock
		if err := rows.Scan(
			&stock.SKU,
			&stock.OnHand,
			&stock.Reserved,
			&stock.UpdatedAt,
		); err != nil {
			return nil, apperror.FromDB(err, "Failed to scan stock row", apperror.CodeInternalError)
		}
		stocks = append(stocks, &stock)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.FromDB(err, "Failed to iterate stock rows", apperror.CodeInternalError)
	}

	return stocks, nil
}

// Create inserts a stock into the database.
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
		return nil, apperror.FromDB(err, "Failed to create stock", apperror.CodeInternalError)
	}

	stock.UpdatedAt = updatedAt

	return stock, nil
}

// Update updates stock fields for a given SKU.
func (sr *StockRepository) Update(
	ctx context.Context,
	request apimodel.StockRequest,
	uow *uow.UnitOfWork,
) (*model.Stock, error) {
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
		return nil, apperror.New(apperror.CodeValidationError, "No updatable fields provided")
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
		return nil, apperror.FromDB(err, "Failed to update stock quantity", apperror.CodeInternalError)
	}

	return &stock, nil
}

// Delete deletes a stock by its SKU.
func (sr *StockRepository) Delete(ctx context.Context, sku string) error {
	query := `DELETE FROM stock WHERE sku = $1 RETURNING sku`

	var deletedSku string
	err := sr.db.QueryRowContext(ctx, query, sku).Scan(&deletedSku)

	if err != nil {
		return apperror.FromDB(err, "Failed to delete stock", apperror.CodeInternalError)
	}

	return nil
}

// Count returns the number of stocks records in the database.
func (sr *StockRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM stock"

	err := sr.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, apperror.FromDB(err, "Failed to count stock records", apperror.CodeInternalError)
	}

	return count, nil
}
