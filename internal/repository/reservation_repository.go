package repository

import (
	"context"
	"github.com/google/uuid"
	"go-inventory-reservations/internal/model"
	"time"
)

// ReservationRepository defines operations for reservation management
type ReservationRepository interface {
	Create(ctx context.Context, reservation *model.Reservation, items []*model.ReservationItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error)
	GetByQuoteID(ctx context.Context, quoteID string) (*model.Reservation, error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Reservation, error)
	GetItems(ctx context.Context, reservationID uuid.UUID) ([]*model.ReservationItem, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, version int) error
	UpdateOrderID(ctx context.Context, id uuid.UUID, orderID string, version int) error
	ListExpired(ctx context.Context, before time.Time) ([]*model.Reservation, error)
}
