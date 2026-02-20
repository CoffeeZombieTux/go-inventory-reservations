package service

import (
	"crypto/sha256"
	"encoding/hex"
	apimodel "go-inventory-reservations/internal/model/api"
	"sort"
	"strconv"
	"strings"
)

// BuildReservationItemsHashFromRequests builds a deterministic hash for active reservation items.
// It normalizes duplicate SKUs, ignores non-positive quantities, and sorts by SKU.
func BuildReservationItemsHashFromRequests(items []apimodel.ReservationItemRequest) *string {
	normalized := make(map[string]int)
	for _, item := range items {
		if item.Quantity <= 0 {
			continue
		}
		normalized[item.SKU] += item.Quantity
	}

	if len(normalized) == 0 {
		return nil
	}

	skus := make([]string, 0, len(normalized))
	for sku := range normalized {
		skus = append(skus, sku)
	}
	sort.Strings(skus)

	var b strings.Builder
	for _, sku := range skus {
		b.WriteString(sku)
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(normalized[sku]))
		b.WriteByte(';')
	}

	sum := sha256.Sum256([]byte(b.String()))
	hash := hex.EncodeToString(sum[:])
	return &hash
}
