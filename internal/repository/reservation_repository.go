package repository

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"go-inventory-reservations/internal/model"
	"time"
)

// ReservationRepository defines operations for reservation management
type ReservationRepositoryInterface interface {
	Create(ctx context.Context, reservation *model.Reservation, items []*model.ReservationItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error)
	GetByQuoteID(ctx context.Context, quoteID string) (*model.Reservation, error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Reservation, error)
	GetItems(ctx context.Context, reservationID uuid.UUID) ([]*model.ReservationItem, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, version int) error
	UpdateOrderID(ctx context.Context, id uuid.UUID, orderID string, version int) error
	ListExpired(ctx context.Context, before time.Time) ([]*model.Reservation, error)
}

type ReservationRepository struct {
	db *sql.DB
}

func (r ReservationRepository) Create(ctx context.Context, reservation *model.Reservation, items []*model.ReservationItem) error {
	//TODO implement me
	panic("implement me")
}

func (r ReservationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	//TODO implement me
	panic("implement me")
}

func (r ReservationRepository) GetByQuoteID(ctx context.Context, quoteID string) (*model.Reservation, error) {
	//TODO implement me
	panic("implement me")
}

func (r ReservationRepository) GetByOrderID(ctx context.Context, orderID string) (*model.Reservation, error) {
	//TODO implement me
	panic("implement me")
}

func (r ReservationRepository) GetItems(ctx context.Context, reservationID uuid.UUID) ([]*model.ReservationItem, error) {
	//TODO implement me
	panic("implement me")
}

func (r ReservationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, version int) error {
	//TODO implement me
	panic("implement me")
}

func (r ReservationRepository) UpdateOrderID(ctx context.Context, id uuid.UUID, orderID string, version int) error {
	//TODO implement me
	panic("implement me")
}

func (r ReservationRepository) ListExpired(ctx context.Context, before time.Time) ([]*model.Reservation, error) {
	//TODO implement me
	panic("implement me")
}

func NewReservationRepository(db *sql.DB) ReservationRepositoryInterface {
	return &ReservationRepository{
		db: db,
	}
}
