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

	"github.com/google/uuid"
)

var ErrReservationVersionConflict = errors.New("reservation version conflict")

// ReservationRepositoryInterface defines operations for reservation management
type ReservationRepositoryInterface interface {
	Save(ctx context.Context, reservation *model.Reservation, uow *uow.UnitOfWork) (*model.Reservation, error)
	GetById(ctx context.Context, id uuid.UUID) (*model.Reservation, error)
	GetByQuoteId(ctx context.Context, quoteId string) (*model.Reservation, error)
	GetByOrderId(ctx context.Context, orderId string) (*model.Reservation, error)
	SelectReservationsByQuery(
		ctx context.Context,
		query apimodel.ReservationsQuery,
		limit int,
	) ([]*model.Reservation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ReservationRepository is a repository for reservation management.
type ReservationRepository struct {
	db *sql.DB
}

// NewReservationRepository creates a new ReservationRepository instance.
func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// Save inserts a reservation into the database or updates it if it already exists.
func (r *ReservationRepository) Save(
	ctx context.Context,
	reservation *model.Reservation,
	uow *uow.UnitOfWork,
) (*model.Reservation, error) {
	var exec SQLExecutor
	if uow != nil && uow.GetTransaction() != nil {
		exec = uow.GetTransaction()
	} else {
		exec = r.db
	}

	if reservation.Version == 0 {
		return r.createReservation(ctx, reservation, exec)
	}

	return r.updateReservationWithVersion(ctx, reservation, exec)
}

// createReservation inserts a new reservation into the database.
func (r *ReservationRepository) createReservation(
	ctx context.Context,
	reservation *model.Reservation,
	exec SQLExecutor,
) (*model.Reservation, error) {
	query := `
		INSERT INTO reservations (reservation_id, status, quote_id, order_id, expires_at, items_hash, version)
		VALUES ($1, $2, $3, $4, $5, $6, 1)
		RETURNING reservation_id, created_at, updated_at, version
	`

	err := exec.QueryRowContext(
		ctx,
		query,
		reservation.ReservationId,
		reservation.Status,
		reservation.QuoteId,
		reservation.OrderId,
		reservation.ExpiresAt,
		reservation.ItemsHash,
	).Scan(&reservation.ReservationId, &reservation.CreatedAt, &reservation.UpdatedAt, &reservation.Version)
	if err != nil {
		return nil, fmt.Errorf("save reservation create failed: %w", err)
	}

	return reservation, nil
}

// updateReservationWithVersion updates a reservation with the given version.
func (r *ReservationRepository) updateReservationWithVersion(
	ctx context.Context,
	reservation *model.Reservation,
	exec SQLExecutor,
) (*model.Reservation, error) {
	query := `
		UPDATE reservations
		SET status = $2,
		    quote_id = $3,
		    order_id = $4,
		    expires_at = $5,
		    items_hash = $6,
		    version = version + 1,
		    updated_at = NOW()
		WHERE reservation_id = $1 AND version = $7
		RETURNING reservation_id, created_at, updated_at, version
	`

	err := exec.QueryRowContext(
		ctx,
		query,
		reservation.ReservationId,
		reservation.Status,
		reservation.QuoteId,
		reservation.OrderId,
		reservation.ExpiresAt,
		reservation.ItemsHash,
		reservation.Version,
	).Scan(&reservation.ReservationId, &reservation.CreatedAt, &reservation.UpdatedAt, &reservation.Version)
	if err == nil {
		return reservation, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf(
			"%w: reservation_id=%s expected_version=%d",
			ErrReservationVersionConflict,
			reservation.ReservationId,
			reservation.Version,
		)
	}

	return nil, fmt.Errorf("save reservation update failed: %w", err)
}

// GetById returns a reservation by its reservationID.
func (r *ReservationRepository) GetById(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	query := `SELECT * FROM reservations WHERE reservation_id = $1`
	res, err := r.getReservation(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("get reservation by id: %w", err)
	}
	return res, nil
}

// GetByQuoteId returns a reservation by its quoteID.
func (r *ReservationRepository) GetByQuoteId(ctx context.Context, quoteId string) (*model.Reservation, error) {
	query := `SELECT * FROM reservations WHERE quote_id = $1`
	res, err := r.getReservation(ctx, query, quoteId)
	if err != nil {
		return nil, fmt.Errorf("get reservation by quote_id failed with error: %w", err)
	}
	return res, nil
}

// GetByOrderId returns a reservation by its orderID.
func (r *ReservationRepository) GetByOrderId(ctx context.Context, orderId string) (*model.Reservation, error) {
	query := `SELECT * FROM reservations WHERE order_id = $1`
	res, err := r.getReservation(ctx, query, orderId)
	if err != nil {
		return nil, fmt.Errorf("get reservation by order_id failed with error: %w", err)
	}
	return res, nil
}

// Delete deletes a reservation by its reservationID.
func (r *ReservationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reservations WHERE reservation_id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete reservation by id: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete reservation by id: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("reservation not found with id: %s", id)
	}
	return nil
}

// SelectReservationsByQuery returns a list of reservations that match the given query.
// Helper function for crons
func (r *ReservationRepository) SelectReservationsByQuery(
	ctx context.Context,
	query apimodel.ReservationsQuery,
	limit int,
) ([]*model.Reservation, error) {
	sql := `SELECT * FROM reservations`
	var where []string
	var args []interface{}
	argPos := 1

	if query.ExpiresAtGte != nil {
		where = append(where, fmt.Sprintf("expires_at <= $%d", argPos))
		args = append(args, *query.ExpiresAtGte)
		argPos++
		fmt.Printf("ExpiresAtGte: %v\n", *query.ExpiresAtGte)
	}
	if query.UpdatedAtLt != nil {
		where = append(where, fmt.Sprintf("updated_at < $%d", argPos))
		args = append(args, *query.UpdatedAtLt)
		argPos++
	}
	if len(query.Statuses) > 0 {
		placeholders := make([]string, len(query.Statuses))
		for i, status := range query.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argPos)
			args = append(args, status)
			argPos++
		}
		where = append(where, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(where) > 0 {
		sql += " WHERE " + strings.Join(where, " AND ")
	}
	sql += fmt.Sprintf(" ORDER BY updated_at ASC LIMIT %d", limit)

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []*model.Reservation
	for rows.Next() {
		var res model.Reservation
		if err := rows.Scan(
			&res.ReservationId,
			&res.Status,
			&res.QuoteId,
			&res.OrderId,
			&res.ExpiresAt,
			&res.CreatedAt,
			&res.UpdatedAt,
			&res.ItemsHash,
			&res.Version,
		); err != nil {
			return nil, err
		}
		reservations = append(reservations, &res)
	}
	return reservations, rows.Err()
}

// getReservation returns a reservation by the given query.
func (r *ReservationRepository) getReservation(
	ctx context.Context,
	query string,
	args ...any,
) (*model.Reservation, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	res := &model.Reservation{}
	err := row.Scan(
		&res.ReservationId,
		&res.Status,
		&res.QuoteId,
		&res.OrderId,
		&res.ExpiresAt,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.ItemsHash,
		&res.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("no reservation found for query: %s", query) // Not found
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan reservation: %w", err)
	}

	return res, nil
}
