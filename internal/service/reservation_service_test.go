package service

import (
	"context"
	"errors"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/repository"
	"go-inventory-reservations/internal/uow"
	"testing"
	"time"

	"github.com/google/uuid"
)

type reservationRepoStub struct {
	getByIDResp     *model.Reservation
	getByIDErr      error
	getByIDForUp    *model.Reservation
	getByIDForUpErr error

	getByQuoteResp *model.Reservation
	getByQuoteErr  error
	getByOrderResp *model.Reservation
	getByOrderErr  error

	saveCalls int
	lastSaved *model.Reservation
	saveErr   error

	selectResp  []*model.Reservation
	selectErr   error
	selectLimit int
	selectQuery apimodel.ReservationsQuery

	deleteCalls int
	deleteErr   error
}

func (r *reservationRepoStub) Save(ctx context.Context, reservation *model.Reservation, u *uow.UnitOfWork) (*model.Reservation, error) {
	r.saveCalls++
	cp := *reservation
	r.lastSaved = &cp
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	if reservation.Version == 0 {
		reservation.Version = 1
	} else {
		reservation.Version++
	}
	return reservation, nil
}

func (r *reservationRepoStub) GetById(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	return r.getByIDResp, nil
}

func (r *reservationRepoStub) GetByIdForUpdate(ctx context.Context, id uuid.UUID, u *uow.UnitOfWork) (*model.Reservation, error) {
	if r.getByIDForUpErr != nil {
		return nil, r.getByIDForUpErr
	}
	return r.getByIDForUp, nil
}

func (r *reservationRepoStub) GetByQuoteId(ctx context.Context, quoteId string) (*model.Reservation, error) {
	if r.getByQuoteErr != nil {
		return nil, r.getByQuoteErr
	}
	return r.getByQuoteResp, nil
}

func (r *reservationRepoStub) GetByOrderId(ctx context.Context, orderId string) (*model.Reservation, error) {
	if r.getByOrderErr != nil {
		return nil, r.getByOrderErr
	}
	return r.getByOrderResp, nil
}

func (r *reservationRepoStub) SelectReservationsByQuery(
	ctx context.Context,
	query apimodel.ReservationsQuery,
	limit int,
) ([]*model.Reservation, error) {
	if r.selectErr != nil {
		return nil, r.selectErr
	}
	r.selectLimit = limit
	r.selectQuery = query
	return r.selectResp, nil
}

func (r *reservationRepoStub) Delete(ctx context.Context, id uuid.UUID) error {
	r.deleteCalls++
	return r.deleteErr
}

func newReservationServiceWithRepo(repo repository.ReservationRepositoryInterface) ReservationService {
	cfg := &config.Config{
		QuoteExpirationSettings: config.QuoteExpirationSettings{
			QuoteExpirationMinutes: 15,
			Limit:                  10,
		},
		ArchiveSettings: config.ArchiveSettings{
			ArchiveReservationsAfterDays: 30,
			Limit:                        50,
		},
	}
	return ReservationService{
		repo:   repo,
		config: cfg,
	}
}

func TestGetToBeExpiredReservations_UsesPendingStatusAndLimit(t *testing.T) {
	repo := &reservationRepoStub{selectResp: []*model.Reservation{}}
	svc := newReservationServiceWithRepo(repo)

	_, err := svc.GetToBeExpiredReservations(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.selectLimit != 10 {
		t.Fatalf("expected limit 10, got %d", repo.selectLimit)
	}
	if len(repo.selectQuery.Statuses) != 1 || repo.selectQuery.Statuses[0] != statusPending {
		t.Fatalf("unexpected statuses query: %+v", repo.selectQuery.Statuses)
	}
	if repo.selectQuery.ExpiresAtGte == nil {
		t.Fatalf("expected expires_at filter")
	}
}

func TestAttachOrder_Success(t *testing.T) {
	id := uuid.New()
	res := &model.Reservation{ReservationId: id, Status: statusPending}
	repo := &reservationRepoStub{getByIDResp: res}
	svc := newReservationServiceWithRepo(repo)

	err := svc.AttachOrder(context.Background(), apimodel.AttachOrderRequest{
		ReservationId: &id,
		OrderId:       "ORD-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saveCalls != 1 {
		t.Fatalf("expected one save call, got %d", repo.saveCalls)
	}
	if repo.lastSaved.Status != statusReserved {
		t.Fatalf("expected RESERVED status, got %s", repo.lastSaved.Status)
	}
	if repo.lastSaved.OrderId == nil || *repo.lastSaved.OrderId != "ORD-1" {
		t.Fatalf("expected order id ORD-1, got %+v", repo.lastSaved.OrderId)
	}
	if repo.lastSaved.ExpiresAt != nil {
		t.Fatalf("expected expires_at to be nil")
	}
}

func TestAttachOrder_InvalidStatus(t *testing.T) {
	id := uuid.New()
	repo := &reservationRepoStub{
		getByIDResp: &model.Reservation{ReservationId: id, Status: statusCommitted},
	}
	svc := newReservationServiceWithRepo(repo)

	err := svc.AttachOrder(context.Background(), apimodel.AttachOrderRequest{
		ReservationId: &id,
		OrderId:       "ORD-1",
	})
	if err == nil {
		t.Fatalf("expected status validation error")
	}
	if repo.saveCalls != 0 {
		t.Fatalf("save should not be called")
	}
}

func TestCreateReservationHelper_SetsDefaults(t *testing.T) {
	repo := &reservationRepoStub{}
	svc := newReservationServiceWithRepo(repo)

	res, err := svc.CreateReservationHelper(context.Background(), apimodel.CreateReservationRequest{
		QuoteId: "Q-1",
		Items: []apimodel.ReservationItemRequest{
			{SKU: "SKU-1", Quantity: 2},
		},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != statusPending {
		t.Fatalf("expected PENDING, got %s", res.Status)
	}
	if res.ExpiresAt == nil {
		t.Fatalf("expected expires_at to be set")
	}
	if res.ItemsHash == nil {
		t.Fatalf("expected items hash to be set")
	}
}

func TestUpdateReservationHelper_NoChangeReturnsOriginal(t *testing.T) {
	id := uuid.New()
	hash := BuildReservationItemsHashFromRequests([]apimodel.ReservationItemRequest{{SKU: "SKU-1", Quantity: 2}})
	repo := &reservationRepoStub{
		getByIDForUp: &model.Reservation{
			ReservationId: id,
			Status:        statusPending,
			QuoteId:       "Q-1",
			ItemsHash:     hash,
			Version:       1,
		},
	}
	svc := newReservationServiceWithRepo(repo)

	res, err := svc.UpdateReservationHelper(context.Background(), apimodel.UpdateReservationRequest{
		ReservationId: &id,
		QuoteId:       "Q-1",
		Items:         []apimodel.ReservationItemRequest{{SKU: "SKU-1", Quantity: 2}},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ReservationId != id {
		t.Fatalf("unexpected reservation returned: %+v", res)
	}
	if repo.saveCalls != 0 {
		t.Fatalf("expected no save when no changes")
	}
}

func TestCommitReservationHelper_Errors(t *testing.T) {
	id := uuid.New()
	repo := &reservationRepoStub{
		getByIDForUp: &model.Reservation{
			ReservationId: id,
			Status:        statusReserved,
			OrderId:       nil,
		},
	}
	svc := newReservationServiceWithRepo(repo)

	_, err := svc.CommitReservationHelper(context.Background(), apimodel.CommitReservationRequest{
		ReservationId: &id,
		OrderId:       "ORD-1",
	}, nil)
	if err == nil {
		t.Fatalf("expected error when order is not attached")
	}

	order := "ORD-2"
	repo.getByIDForUp = &model.Reservation{
		ReservationId: id,
		Status:        statusReserved,
		OrderId:       &order,
	}
	_, err = svc.CommitReservationHelper(context.Background(), apimodel.CommitReservationRequest{
		ReservationId: &id,
		OrderId:       "ORD-1",
	}, nil)
	if err == nil {
		t.Fatalf("expected mismatch error")
	}
}

func TestCommitReservationHelper_Success(t *testing.T) {
	id := uuid.New()
	order := "ORD-1"
	repo := &reservationRepoStub{
		getByIDForUp: &model.Reservation{
			ReservationId: id,
			Status:        statusReserved,
			OrderId:       &order,
			Version:       2,
			ItemsHash:     ptrString("hash"),
		},
	}
	svc := newReservationServiceWithRepo(repo)

	res, err := svc.CommitReservationHelper(context.Background(), apimodel.CommitReservationRequest{
		ReservationId: &id,
		OrderId:       "ORD-1",
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != statusCommitted {
		t.Fatalf("expected committed status, got %s", res.Status)
	}
	if res.ItemsHash != nil {
		t.Fatalf("expected items hash to be nil")
	}
}

func TestReleaseReservationHelper_Success(t *testing.T) {
	id := uuid.New()
	repo := &reservationRepoStub{
		getByIDForUp: &model.Reservation{
			ReservationId: id,
			Status:        statusPending,
			ItemsHash:     ptrString("h"),
		},
	}
	svc := newReservationServiceWithRepo(repo)

	res, err := svc.ReleaseReservationHelper(context.Background(), id, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != statusReleased {
		t.Fatalf("expected released status, got %s", res.Status)
	}
	if res.ItemsHash != nil {
		t.Fatalf("expected items hash nil")
	}
}

func TestRevertReservationHelper_ErrorsAndSuccess(t *testing.T) {
	id := uuid.New()
	order := "ORD-1"
	repo := &reservationRepoStub{
		getByIDForUp: &model.Reservation{
			ReservationId: id,
			Status:        statusPending,
			OrderId:       &order,
		},
	}
	svc := newReservationServiceWithRepo(repo)

	_, err := svc.RevertReservationHelper(context.Background(), apimodel.RevertReservationRequest{
		ReservationId: &id,
		OrderId:       order,
	}, nil)
	if err == nil {
		t.Fatalf("expected status validation error")
	}

	repo.getByIDForUp = &model.Reservation{
		ReservationId: id,
		Status:        statusCommitted,
		OrderId:       nil,
	}
	_, err = svc.RevertReservationHelper(context.Background(), apimodel.RevertReservationRequest{
		ReservationId: &id,
		OrderId:       order,
	}, nil)
	if err == nil {
		t.Fatalf("expected no-order error")
	}

	repo.getByIDForUp = &model.Reservation{
		ReservationId: id,
		Status:        statusCommitted,
		OrderId:       &order,
	}
	res, err := svc.RevertReservationHelper(context.Background(), apimodel.RevertReservationRequest{
		ReservationId: &id,
		OrderId:       order,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != statusReverted {
		t.Fatalf("expected reverted, got %s", res.Status)
	}
}

func TestExpireReservationHelper(t *testing.T) {
	repo := &reservationRepoStub{}
	svc := newReservationServiceWithRepo(repo)
	res := &model.Reservation{Status: statusPending, ItemsHash: ptrString("h")}

	err := svc.ExpireReservationHelper(context.Background(), res, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != statusExpired || res.ItemsHash != nil {
		t.Fatalf("unexpected reservation after expire: %+v", res)
	}
}

func TestArchiveReservations(t *testing.T) {
	repo := &reservationRepoStub{
		selectResp: []*model.Reservation{
			{ReservationId: uuid.New()},
			{ReservationId: uuid.New()},
		},
	}
	svc := newReservationServiceWithRepo(repo)

	n, err := svc.ArchiveReservations(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 archived, got %d", n)
	}
	if repo.deleteCalls != 2 {
		t.Fatalf("expected 2 deletes, got %d", repo.deleteCalls)
	}
	if repo.selectQuery.UpdatedAtLt == nil {
		t.Fatalf("expected updated_at filter")
	}
}

func TestHelperFunctions(t *testing.T) {
	if err := checkAvailableStatuses(model.Reservation{Status: statusPending}, []string{statusPending}); err != nil {
		t.Fatalf("expected allowed status")
	}
	if err := checkAvailableStatuses(model.Reservation{Status: statusCommitted}, []string{statusPending}); err == nil {
		t.Fatalf("expected disallowed status error")
	}

	a := ptrString("x")
	b := ptrString("x")
	c := ptrString("y")
	if !itemsHashEqual(a, b) {
		t.Fatalf("expected equal hashes")
	}
	if itemsHashEqual(a, c) {
		t.Fatalf("expected non-equal hashes")
	}
	if !itemsHashEqual(nil, nil) {
		t.Fatalf("expected nil hashes equal")
	}

	if !IsReservationVersionConflict(errors.Join(errors.New("wrap"), repository.ErrReservationVersionConflict)) {
		t.Fatalf("expected conflict detection to be true")
	}
	if IsReservationVersionConflict(errors.New("other")) {
		t.Fatalf("expected conflict detection to be false")
	}
}

func TestGetExpiresAt(t *testing.T) {
	repo := &reservationRepoStub{}
	svc := newReservationServiceWithRepo(repo)
	before := time.Now().Add(14 * time.Minute)
	expires := svc.getExpiresAt()
	after := time.Now().Add(16 * time.Minute)
	if expires.Before(before) || expires.After(after) {
		t.Fatalf("unexpected expires_at: %v", expires)
	}
}

func ptrString(v string) *string {
	return &v
}
