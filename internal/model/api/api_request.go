package api_model

import (
	"github.com/google/uuid"
	"time"
)

// StockRequest represents a request to reserve stock items
type StockRequest struct {
	SKU      string `json:"sku" binding:"required"`
	OnHand   *int   `json:"on_hand"`
	Reserved *int   `json:"reserved"`
}

// PaginationParams represents pagination parameters for a request
type PaginationParams struct {
	Limit  int
	Offset int
}

// NewPaginationParams creates a new PaginationParams instance
func NewPaginationParams(requestedLimit, requestedOffset int) PaginationParams {
	const (
		DefaultLimit = 50
		MaxLimit     = 10000
		MinLimit     = 1
	)

	// Apply constraints to the limit
	limit := DefaultLimit
	if requestedLimit >= MinLimit {
		limit = requestedLimit
		if limit > MaxLimit {
			// Silent Adjustment
			limit = MaxLimit
		}
	}

	// Ensure the offset is not negative
	offset := 0
	if requestedOffset > 0 {
		offset = requestedOffset
	}

	return PaginationParams{
		Limit:  limit,
		Offset: offset,
	}
}

// CreateReservationRequest represents a request to create a reservation
type CreateReservationRequest struct {
	QuoteId string                   `json:"quote_id" binding:"required"`
	Items   []ReservationItemRequest `json:"items" binding:"required,gt=0"`
}

// ReservationItemRequest represents a request to reserve a single item
type ReservationItemRequest struct {
	SKU      string `json:"sku" binding:"required"`
	Quantity int    `json:"qty" binding:"required,gte=0"`
}

// UpdateReservationRequest represents a request to update a reservation
type UpdateReservationRequest struct {
	ReservationId *uuid.UUID               `json:"reservation_id" binding:"required"`
	QuoteId       string                   `json:"quote_id" binding:"required"`
	Items         []ReservationItemRequest `json:"items" binding:"required,gt=0"`
}

// CommitReservationRequest represents a request to commit a reservation
type CommitReservationRequest struct {
	ReservationId *uuid.UUID `json:"reservation_id" binding:"required"`
	OrderId       string     `json:"order_id" binding:"required"`
}

// AttachOrderRequest represents a request to attach an order to a reservation
type AttachOrderRequest struct {
	ReservationId *uuid.UUID `json:"reservation_id" binding:"required"`
	OrderId       string     `json:"order_id" binding:"required"`
}

// RevertReservationRequest represents a request to revert a reservation
type RevertReservationRequest struct {
	ReservationId *uuid.UUID `json:"reservation_id" binding:"required"`
	OrderId       string     `json:"order_id" binding:"required"`
}

// ReservationsQuery represents a query to retrieve reservations
type ReservationsQuery struct {
	ExpiresAtGte *time.Time
	Statuses     []string
	UpdatedAtLt  *time.Time
}
