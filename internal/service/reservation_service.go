package service

import (
	"context"
	"github.com/google/uuid"
	"go-inventory-reservations/internal/model"
	"go-inventory-reservations/internal/repository"
	"time"
)

// ReservationService defines business operations for reservation management
type ReservationServiceInterface interface {
	CreateReservation(ctx context.Context, quoteID string, items []model.ReservationItem, expiresAt time.Time) (*model.Reservation, error)
	GetReservation(ctx context.Context, id uuid.UUID) (*model.Reservation, []*model.ReservationItem, error)
	CommitReservation(ctx context.Context, id uuid.UUID, orderID string) error
	ReleaseReservation(ctx context.Context, id uuid.UUID) error
	ProcessExpiredReservations(ctx context.Context) (int, error)
}

type ReservationService struct {
	repo repository.ReservationRepositoryInterface
}

func NewReservationService(repo repository.ReservationRepositoryInterface) ReservationServiceInterface {
	return &ReservationService{
		repo: repo,
	}
}

func (rs ReservationService) CreateReservation(ctx context.Context, quoteID string, items []model.ReservationItem, expiresAt time.Time) (*model.Reservation, error) {
	//TODO implement me
	panic("implement me")
}

func (rs ReservationService) GetReservation(ctx context.Context, id uuid.UUID) (*model.Reservation, []*model.ReservationItem, error) {
	//TODO implement me
	panic("implement me")
}

func (rs ReservationService) CommitReservation(ctx context.Context, id uuid.UUID, orderID string) error {
	//TODO implement me
	panic("implement me")
}

func (rs ReservationService) ReleaseReservation(ctx context.Context, id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (rs ReservationService) ProcessExpiredReservations(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}
