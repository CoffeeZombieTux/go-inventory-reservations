package service

import (
	apimodel "go-inventory-reservations/internal/model/api"
	"testing"
)

// TestBuildReservationItemsHashFromRequests_OrderIndependent verifies hash stability for reordered items.
func TestBuildReservationItemsHashFromRequests_OrderIndependent(t *testing.T) {
	itemsA := []apimodel.ReservationItemRequest{
		{SKU: "SKU-2", Quantity: 1},
		{SKU: "SKU-1", Quantity: 3},
	}
	itemsB := []apimodel.ReservationItemRequest{
		{SKU: "SKU-1", Quantity: 3},
		{SKU: "SKU-2", Quantity: 1},
	}

	hashA := BuildReservationItemsHashFromRequests(itemsA)
	hashB := BuildReservationItemsHashFromRequests(itemsB)

	if hashA == nil || hashB == nil {
		t.Fatalf("expected non-nil hashes")
	}
	if *hashA != *hashB {
		t.Fatalf("expected equal hashes, got %s and %s", *hashA, *hashB)
	}
}

// TestBuildReservationItemsHashFromRequests_DuplicateSKUAggregated verifies duplicate SKU quantities are aggregated.
func TestBuildReservationItemsHashFromRequests_DuplicateSKUAggregated(t *testing.T) {
	itemsA := []apimodel.ReservationItemRequest{
		{SKU: "SKU-1", Quantity: 1},
		{SKU: "SKU-1", Quantity: 2},
		{SKU: "SKU-2", Quantity: 4},
	}
	itemsB := []apimodel.ReservationItemRequest{
		{SKU: "SKU-1", Quantity: 3},
		{SKU: "SKU-2", Quantity: 4},
	}

	hashA := BuildReservationItemsHashFromRequests(itemsA)
	hashB := BuildReservationItemsHashFromRequests(itemsB)

	if hashA == nil || hashB == nil {
		t.Fatalf("expected non-nil hashes")
	}
	if *hashA != *hashB {
		t.Fatalf("expected equal hashes, got %s and %s", *hashA, *hashB)
	}
}

// TestBuildReservationItemsHashFromRequests_IgnoresNonPositive verifies non-positive quantities are ignored.
func TestBuildReservationItemsHashFromRequests_IgnoresNonPositive(t *testing.T) {
	items := []apimodel.ReservationItemRequest{
		{SKU: "SKU-1", Quantity: 0},
		{SKU: "SKU-2", Quantity: -2},
	}

	hash := BuildReservationItemsHashFromRequests(items)
	if hash != nil {
		t.Fatalf("expected nil hash for non-positive-only items, got %s", *hash)
	}
}
