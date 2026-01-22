package model

import (
	"github.com/google/uuid"
)

// ReservationItem represents an item within a reservation
type ReservationItem struct {
	ReservationID uuid.UUID `db:"reservation_id" json:"reservation_id"` // Reference to the parent reservation
	SKU           string    `db:"sku" json:"sku"`                       // Product SKU
	Qty           int       `db:"qty" json:"qty"`                       // Quantity of items being reserved
}
