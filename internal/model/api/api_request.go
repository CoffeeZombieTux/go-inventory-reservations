package api_model

type StockRequest struct {
	SKU      string `json:"sku" binding:"required"`
	Quantity int    `json:"qty" binding:"required,gte=0"`
}

type PaginationParams struct {
	Limit  int
	Offset int
}

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
