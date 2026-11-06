package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
)

func TestEstimateJobCost(t *testing.T) {
	rc := api.RateCard{
		MachineRatePerHour: 10,
		FilamentPricePerKg: 25,
		FilamentKgPerHour:  0.1,
	}
	got := db.EstimateJobCost(rc, 3600, 100, 1)
	if got != 12.5 {
		t.Fatalf("got %v want 12.5", got)
	}
	got = db.EstimateJobCost(rc, 3600, 100, 2)
	if got != 25 {
		t.Fatalf("got %v want 25", got)
	}
	got = db.EstimateJobCost(rc, 3600, 0, 1)
	if got != 12.5 {
		t.Fatalf("derived grams: got %v want 12.5", got)
	}
}
