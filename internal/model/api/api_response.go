package api_model

// PaginationResponse represents a response containing pagination metadata
type PaginationResponse struct {
	Limit       int `json:"limit"`
	Offset      int `json:"offset"`
	TotalItems  int `json:"total_items"`
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
}
