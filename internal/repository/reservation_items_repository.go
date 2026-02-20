package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-inventory-reservations/internal/model"
	"go-inventory-reservations/internal/uow"
)

// ReservationItemsRepositoryInterface defines operations for reservation items management
type ReservationItemsRepositoryInterface interface {
	Get(ctx context.Context, reservationId uuid.UUID, sku string, uow *uow.UnitOfWork) (*model.ReservationItem, error)
	FindByReservationId(ctx context.Context, reservationId uuid.UUID, uow *uow.UnitOfWork) (map[string]*model.ReservationItem, error)
	FindActiveBySku(ctx context.Context, sku string, uow *uow.UnitOfWork) (map[string]*model.ReservationItem, error)
	Create(ctx context.Context, item *model.ReservationItem, uow *uow.UnitOfWork) (*model.ReservationItem, error)
	Update(ctx context.Context, item *model.ReservationItem, uow *uow.UnitOfWork) (*model.ReservationItem, error)
	SetIsActive(ctx context.Context, reservationId uuid.UUID, sku string, isActive bool, uow *uow.UnitOfWork) (*model.ReservationItem, error)
	Delete(ctx context.Context, reservationId uuid.UUID, sku string, uow *uow.UnitOfWork) error
}

// ReservationItemsRepository is a repository for reservation items management.
type ReservationItemsRepository struct {
	db *sql.DB
}

// NewReservationItemsRepository creates a new ReservationItemsRepository instance.
func NewReservationItemsRepository(db *sql.DB) *ReservationItemsRepository {
	return &ReservationItemsRepository{db: db}
}

// Get retrieves a reservation item by reservation ID and SKU using the provided context and unit of work if available.
func (rir *ReservationItemsRepository) Get(
	ctx context.Context,
	reservationId uuid.UUID,
	sku string,
	uow *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	var exec SQLExecutor
	if uow != nil && uow.GetTransaction() != nil {
		exec = uow.GetTransaction()
	} else {
		exec = rir.db
	}

	query := `
		SELECT reservation_id, sku, qty, is_active
		FROM reservation_items
		WHERE reservation_id = $1 AND sku = $2
	`
	var item model.ReservationItem
	err := exec.QueryRowContext(ctx, query, reservationId, sku).Scan(&item.ReservationId, &item.SKU, &item.Qty, &item.IsActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation item: %w", err)
	}
	return &item, nil
}

// FindByReservationId retrieves all reservation items for a given reservation ID using the provided context and unit of work if available.
func (rir *ReservationItemsRepository) FindByReservationId(
	ctx context.Context,
	reservationId uuid.UUID,
	uow *uow.UnitOfWork,
) (map[string]*model.ReservationItem, error) {
	var exec SQLExecutor
	if uow != nil && uow.GetTransaction() != nil {
		exec = uow.GetTransaction()
	} else {
		exec = rir.db
	}

	query := `
		SELECT reservation_id, sku, qty, is_active
		FROM reservation_items
		WHERE reservation_id = $1
	`
	rows, err := exec.QueryContext(ctx, query, reservationId)
	if err != nil {
		return nil, fmt.Errorf("failed to query reservation items: %w", err)
	}
	defer rows.Close()

	items := make(map[string]*model.ReservationItem)
	for rows.Next() {
		var item model.ReservationItem
		if err := rows.Scan(&item.ReservationId, &item.SKU, &item.Qty, &item.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan reservation item: %w", err)
		}
		items[item.SKU] = &item
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservation items rows: %w", err)
	}
	return items, nil
}

// FindActiveBySku retrieves active reservation items for a given SKU.
func (rir *ReservationItemsRepository) FindActiveBySku(
	ctx context.Context,
	sku string,
	uow *uow.UnitOfWork,
) (map[string]*model.ReservationItem, error) {
	var exec SQLExecutor
	if uow != nil && uow.GetTransaction() != nil {
		exec = uow.GetTransaction()
	} else {
		exec = rir.db
	}

	query := `
		SELECT reservation_id, sku, qty, is_active
		FROM reservation_items
		WHERE sku = $1 AND is_active = true
	`

	rows, err := exec.QueryContext(ctx, query, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to query reservation items: %w", err)
	}
	defer rows.Close()

	items := make(map[string]*model.ReservationItem)
	for rows.Next() {
		var item model.ReservationItem
		if err := rows.Scan(&item.ReservationId, &item.SKU, &item.Qty, &item.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan reservation item: %w", err)
		}
		items[item.ReservationId.String()] = &item
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservation items rows: %w", err)
	}

	return items, nil
}

// Create inserts a new reservation item into the database using the provided context and unit of work if available.
func (rir *ReservationItemsRepository) Create(
	ctx context.Context,
	item *model.ReservationItem,
	uow *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	var exec SQLExecutor
	if uow != nil && uow.GetTransaction() != nil {
		exec = uow.GetTransaction()
	} else {
		exec = rir.db
	}

	query := `
		INSERT INTO reservation_items (
			reservation_id, sku, qty, is_active
		) VALUES ($1, $2, $3, $4)
	`
	_, err := exec.ExecContext(ctx, query, item.ReservationId, item.SKU, item.Qty, item.IsActive)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation item: %w", err)
	}

	return item, nil
}

// Update updates an existing reservation item in the database using the provided context and unit of work if available.
func (rir *ReservationItemsRepository) Update(
	ctx context.Context,
	item *model.ReservationItem,
	uow *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	var exec SQLExecutor
	if uow != nil && uow.GetTransaction() != nil {
		exec = uow.GetTransaction()
	} else {
		exec = rir.db
	}

	query := `
		UPDATE reservation_items
		SET qty = $2, is_active = $3
		WHERE reservation_id = $1 AND sku = $4
		RETURNING qty, is_active
	`

	err := exec.QueryRowContext(ctx, query, item.ReservationId, item.Qty, item.IsActive, item.SKU).Scan(&item.Qty, &item.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(
				"reservation item not found for ReservationID %s and SKU: %s",
				item.ReservationId,
				item.SKU,
			)
		}
		return nil, fmt.Errorf("failed to update reservation item: %w", err)
	}
	return item, nil
}

// SetIsActive updates reservation item active flag in the database.
func (rir *ReservationItemsRepository) SetIsActive(
	ctx context.Context,
	reservationId uuid.UUID,
	sku string,
	isActive bool,
	uow *uow.UnitOfWork,
) (*model.ReservationItem, error) {
	var exec SQLExecutor
	if uow != nil && uow.GetTransaction() != nil {
		exec = uow.GetTransaction()
	} else {
		exec = rir.db
	}

	query := `
		UPDATE reservation_items
		SET is_active = $3
		WHERE reservation_id = $1 AND sku = $2
		RETURNING reservation_id, sku, qty, is_active
	`

	var item model.ReservationItem
	err := exec.QueryRowContext(ctx, query, reservationId, sku, isActive).
		Scan(&item.ReservationId, &item.SKU, &item.Qty, &item.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("reservation item not found for ReservationID %s and SKU: %s", reservationId, sku)
		}
		return nil, fmt.Errorf("failed to update reservation item active flag: %w", err)
	}
	return &item, nil
}

// Delete deletes a reservation item from the database using the provided context and unit of work if available.
func (rir *ReservationItemsRepository) Delete(
	ctx context.Context,
	reservationId uuid.UUID,
	sku string,
	uow *uow.UnitOfWork,
) error {
	var exec SQLExecutor
	if uow != nil && uow.GetTransaction() != nil {
		exec = uow.GetTransaction()
	} else {
		exec = rir.db
	}
	query := `DELETE FROM reservation_items WHERE reservation_id = $1 AND sku = $2 RETURNING reservation_id, sku`
	var deletedItem model.ReservationItem
	err := exec.QueryRowContext(ctx, query, reservationId, sku).Scan(&deletedItem.ReservationId, &deletedItem.SKU)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("reservation item not found for ReservationID %s and SKU: %s", reservationId, sku)
		}
		return fmt.Errorf("failed to update reservation item: %w", err)
	}
	return nil
}
