package service

import (
	"context"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/model"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/uow"
	"testing"

	"github.com/google/uuid"
)

type reservationItemsRepoStub struct {
	getResp          *model.ReservationItem
	getErr           error
	byReservationMap map[string]*model.ReservationItem
	byReservationErr error
	bySkuMap         map[string]*model.ReservationItem
	bySkuErr         error

	createCalls int
	lastCreated *model.ReservationItem
	createErr   error

	updateCalls int
	lastUpdated *model.ReservationItem
	updateErr   error

	setActiveResp *model.ReservationItem
	setActiveErr  error

	deleteCalls int
	deleteErr   error
}

func (r *reservationItemsRepoStub) Get(ctx context.Context, reservationId uuid.UUID, sku string, u *uow.UnitOfWork) (*model.ReservationItem, error) {
	return r.getResp, r.getErr
}

func (r *reservationItemsRepoStub) FindByReservationId(ctx context.Context, reservationId uuid.UUID, u *uow.UnitOfWork) (map[string]*model.ReservationItem, error) {
	return r.byReservationMap, r.byReservationErr
}

func (r *reservationItemsRepoStub) FindActiveBySku(ctx context.Context, sku string, u *uow.UnitOfWork) (map[string]*model.ReservationItem, error) {
	return r.bySkuMap, r.bySkuErr
}

func (r *reservationItemsRepoStub) Create(ctx context.Context, item *model.ReservationItem, u *uow.UnitOfWork) (*model.ReservationItem, error) {
	r.createCalls++
	cp := *item
	r.lastCreated = &cp
	return item, r.createErr
}

func (r *reservationItemsRepoStub) Update(ctx context.Context, item *model.ReservationItem, u *uow.UnitOfWork) (*model.ReservationItem, error) {
	r.updateCalls++
	cp := *item
	r.lastUpdated = &cp
	return item, r.updateErr
}

func (r *reservationItemsRepoStub) SetIsActive(ctx context.Context, reservationId uuid.UUID, sku string, isActive bool, u *uow.UnitOfWork) (*model.ReservationItem, error) {
	return r.setActiveResp, r.setActiveErr
}

func (r *reservationItemsRepoStub) Delete(ctx context.Context, reservationId uuid.UUID, sku string, u *uow.UnitOfWork) error {
	r.deleteCalls++
	return r.deleteErr
}

func newReservationItemsServiceWithRepo(repo *reservationItemsRepoStub) ReservationItemsService {
	return ReservationItemsService{
		repo: repo,
		config: &config.Config{
			ReservationItemSettings: config.ReservationItemSettings{MaxQuantity: 5},
		},
	}
}

func TestReservationItemsService_Getters(t *testing.T) {
	id := uuid.New()
	item := &model.ReservationItem{ReservationId: id, SKU: "SKU-1", Qty: 2, IsActive: true}
	repo := &reservationItemsRepoStub{
		getResp: item,
		byReservationMap: map[string]*model.ReservationItem{
			"SKU-1": item,
		},
		bySkuMap: map[string]*model.ReservationItem{
			id.String(): item,
		},
	}
	svc := newReservationItemsServiceWithRepo(repo)

	got, err := svc.GetReservationItem(context.Background(), id, "SKU-1", nil)
	if err != nil || got == nil {
		t.Fatalf("expected item, err=%v got=%+v", err, got)
	}

	byRes, err := svc.GetReservationItems(context.Background(), id, nil)
	if err != nil || len(byRes) != 1 {
		t.Fatalf("unexpected GetReservationItems result: err=%v len=%d", err, len(byRes))
	}

	bySku, err := svc.GetActiveReservationItemsBySku(context.Background(), "SKU-1", nil)
	if err != nil || len(bySku) != 1 {
		t.Fatalf("unexpected GetActiveReservationItemsBySku result: err=%v len=%d", err, len(bySku))
	}
}

func TestReservationItemsService_CreateValidationAndSuccess(t *testing.T) {
	id := uuid.New()
	repo := &reservationItemsRepoStub{}
	svc := newReservationItemsServiceWithRepo(repo)

	_, err := svc.CreateReservationItem(context.Background(), apimodel.ReservationItemRequest{
		SKU:      "SKU-1",
		Quantity: 6,
	}, id, nil)
	if err == nil {
		t.Fatalf("expected max quantity validation error")
	}
	if repo.createCalls != 0 {
		t.Fatalf("create should not be called on validation error")
	}

	created, err := svc.CreateReservationItem(context.Background(), apimodel.ReservationItemRequest{
		SKU:      "SKU-1",
		Quantity: 3,
	}, id, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created == nil || !created.IsActive {
		t.Fatalf("expected active created item, got %+v", created)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected one create call, got %d", repo.createCalls)
	}
}

func TestReservationItemsService_UpdateAndDelete(t *testing.T) {
	id := uuid.New()
	repo := &reservationItemsRepoStub{
		getResp: &model.ReservationItem{ReservationId: id, SKU: "SKU-1", Qty: 2, IsActive: true},
		setActiveResp: &model.ReservationItem{
			ReservationId: id,
			SKU:           "SKU-1",
			Qty:           2,
			IsActive:      false,
		},
	}
	svc := newReservationItemsServiceWithRepo(repo)

	updated, err := svc.UpdateReservationItem(context.Background(), apimodel.ReservationItemRequest{
		SKU:      "SKU-1",
		Quantity: 0,
	}, id, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.IsActive {
		t.Fatalf("expected updated item to be inactive")
	}
	if repo.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", repo.updateCalls)
	}

	item, err := svc.SetReservationItemActive(context.Background(), id, "SKU-1", false, nil)
	if err != nil || item == nil || item.IsActive {
		t.Fatalf("unexpected SetReservationItemActive result: err=%v item=%+v", err, item)
	}

	if err := svc.DeleteReservationItem(context.Background(), id, "SKU-1", nil); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if repo.deleteCalls != 1 {
		t.Fatalf("expected one delete call, got %d", repo.deleteCalls)
	}
}
