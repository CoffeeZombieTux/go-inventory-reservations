//go:build integration
// +build integration

package repository

import (
	"context"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/model"
	"go-inventory-reservations/internal/testutil"
	"testing"
)

func TestStockRepository_ResetDBAndCreateStock(t *testing.T) {
	db := testutil.OpenTestDB(t)
	defer db.Close()

	ctx := context.Background()
	testutil.ResetAndMigrateTestDB(t, ctx, db)

	repo := NewStockRepository(db, logger.New("error", "text"))

	created, err := repo.Create(ctx, &model.Stock{
		SKU:      "SKU-TEST-001",
		OnHand:   25,
		Reserved: 0,
	})
	if err != nil {
		t.Fatalf("create stock failed: %v", err)
	}

	if created.SKU != "SKU-TEST-001" || created.OnHand != 25 || created.Reserved != 0 {
		t.Fatalf("unexpected created stock payload: %+v", created)
	}
	if created.UpdatedAt.IsZero() {
		t.Fatalf("expected updated_at to be set")
	}

	stored, err := repo.GetBySku(ctx, "SKU-TEST-001")
	if err != nil {
		t.Fatalf("get stock by sku failed: %v", err)
	}

	if stored.SKU != "SKU-TEST-001" || stored.OnHand != 25 || stored.Reserved != 0 {
		t.Fatalf("unexpected stored stock payload: %+v", stored)
	}
}
