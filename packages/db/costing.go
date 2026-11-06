package db

import "github.com/jjheffernan/BambuFarmer/packages/api"

// EstimateJobCost computes machine + filament cost from rate card and job estimates.
// machine = (est_seconds/3600) * $/hr ; filament = (est_grams/1000) * $/kg
// If est_grams is 0 and kg/hr > 0, filament mass ≈ hours * kg/hr.
func EstimateJobCost(rc api.RateCard, estSeconds int, estGrams float64, copies int) float64 {
	if copies < 1 {
		copies = 1
	}
	hours := float64(estSeconds) / 3600.0
	grams := estGrams
	if grams <= 0 && rc.FilamentKgPerHour > 0 {
		grams = hours * rc.FilamentKgPerHour * 1000.0
	}
	machine := hours * rc.MachineRatePerHour
	filament := (grams / 1000.0) * rc.FilamentPricePerKg
	return (machine + filament + rc.LaborFlat) * float64(copies)
}
