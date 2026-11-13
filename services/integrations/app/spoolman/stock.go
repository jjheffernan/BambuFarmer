package spoolman

import (
	"strconv"
	"strings"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

const defaultLowGrams = 100

// LowGrams parses LOW_STOCK_GRAMS. Empty or invalid uses 100.
func LowGrams(raw string) float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultLowGrams
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil || n <= 0 {
		return defaultLowGrams
	}
	return n
}

// LowStock returns non-archived spools under the remaining-grams threshold.
func LowStock(spools []api.SpoolSummary, threshold float64) []api.SpoolSummary {
	if threshold <= 0 {
		threshold = defaultLowGrams
	}
	out := []api.SpoolSummary{}
	for _, s := range spools {
		if s.Archived {
			continue
		}
		if s.RemainingG < threshold {
			out = append(out, s)
		}
	}
	return out
}
