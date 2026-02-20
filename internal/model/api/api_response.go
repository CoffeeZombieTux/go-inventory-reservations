package api_model

import (
	"go-inventory-reservations/internal/model"
)

// StockResponse represents a stock response
type StockResponse struct {
	SKU       string `json:"sku"`
	OnHand    int    `json:"on_hand"`
	Reserved  int    `json:"reserved"`
	Available int    `json:"available_quantity"`
}

// PaginationResponse represents a response containing pagination metadata
type PaginationResponse struct {
	Limit       int `json:"limit"`
	Offset      int `json:"offset"`
	TotalItems  int `json:"total_items"`
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
}

// ReservationResponse represents a reservation response with items included
type ReservationResponse struct {
	Reservation *model.Reservation                `json:"reservation"`
	Items       map[string]*model.ReservationItem `json:"items"`
}
