package api_model

import "testing"

func TestNewPaginationParams_Defaults(t *testing.T) {
	params := NewPaginationParams(0, 0)
	if params.Limit != 50 {
		t.Fatalf("expected default limit 50, got %d", params.Limit)
	}
	if params.Offset != 0 {
		t.Fatalf("expected default offset 0, got %d", params.Offset)
	}
}

func TestNewPaginationParams_MaxLimitCap(t *testing.T) {
	params := NewPaginationParams(20000, 1)
	if params.Limit != 10000 {
		t.Fatalf("expected capped limit 10000, got %d", params.Limit)
	}
	if params.Offset != 1 {
		t.Fatalf("expected offset 1, got %d", params.Offset)
	}
}

func TestNewPaginationParams_NegativeOffset(t *testing.T) {
	params := NewPaginationParams(10, -99)
	if params.Offset != 0 {
		t.Fatalf("expected non-negative offset 0, got %d", params.Offset)
	}
}
