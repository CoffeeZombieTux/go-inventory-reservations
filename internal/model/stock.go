package model

import (
	"time"
)

// Stock represents an inventory item record
type Stock struct {
	SKU       string    `db:"sku" json:"sku"`               // Primary key for the item
	OnHand    int       `db:"on_hand" json:"on_hand"`       // Available quantity in stock
	Reserved  int       `db:"reserved" json:"reserved"`     // Currently reserved quantity
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"` // Last update timestamp
}
