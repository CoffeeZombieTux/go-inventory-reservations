package service

import (
	"context"
	"github.com/google/uuid"
	"go-inventory-reservations/internal/model"
	"time"
)

// ReservationService defines business operations for reservation management
type ReservationService interface {
	CreateReservation(ctx context.Context, quoteID string, items []model.ReservationItem, expiresAt time.Time) (*model.Reservation, error)
	GetReservation(ctx context.Context, id uuid.UUID) (*model.Reservation, []*model.ReservationItem, error)
	CommitReservation(ctx context.Context, id uuid.UUID, orderID string) error
	ReleaseReservation(ctx context.Context, id uuid.UUID) error
	ProcessExpiredReservations(ctx context.Context) (int, error)
}
