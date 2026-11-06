package db

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

func TestEstimateJobCost(t *testing.T) {
	rc := api.RateCard{
		MachineRatePerHour: 10,  // $/hr
		FilamentPricePerKg: 25,  // $/kg
		FilamentKgPerHour:  0.1, // unused when grams set
	}
	// 1 hour, 100g → 10 + 2.5 = 12.5
	got := EstimateJobCost(rc, 3600, 100, 1)
	if got != 12.5 {
		t.Fatalf("got %v want 12.5", got)
	}
	// copies
	got = EstimateJobCost(rc, 3600, 100, 2)
	if got != 25 {
		t.Fatalf("got %v want 25", got)
	}
	// derive grams from kg/hr when grams=0: 1h * 0.1kg = 100g → same 12.5
	got = EstimateJobCost(rc, 3600, 0, 1)
	if got != 12.5 {
		t.Fatalf("derived grams: got %v want 12.5", got)
	}
}
