package api_model

import "go-inventory-reservations/internal/model"

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
