package model

import (
	"github.com/google/uuid"
	"time"
)

// Reservation represents a reservation of inventory items
type Reservation struct {
	ReservationId uuid.UUID  `db:"reservation_id" json:"reservation_id"` // Primary key for the reservation
	Status        string     `db:"status" json:"status"`                 // Status of the reservation (PENDING, COMMITTED, RELEASED, EXPIRED)
	QuoteId       *string    `db:"quote_id" json:"quote_id"`             // Quote Id reference (can be NULL)
	OrderId       *string    `db:"order_id" json:"order_id"`             // Order Id reference (can be NULL)
	ExpiresAt     *time.Time `db:"expires_at" json:"expires_at"`         // Expiration timestamp
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`         // Creation timestamp
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`         // Last update timestamp
	ItemsHash     *string    `db:"items_hash" json:"items_hash"`         // Hash of the items (can be NULL)
	Version       int        `db:"version" json:"version"`               // Version for optimistic locking
}
