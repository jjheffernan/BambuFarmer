package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// GetRateCard returns the site rate card (creates default row if missing via seed).
func GetRateCard(ctx context.Context, pool *pgxpool.Pool, siteID string) (api.RateCard, error) {
	var rc api.RateCard
	err := pool.QueryRow(ctx, `
		SELECT currency, machine_rate_per_hour, filament_price_per_kg, filament_kg_per_hour,
		       labor_flat, failure_surcharge_pct
		FROM rate_cards WHERE site_id=$1`, siteID).Scan(
		&rc.Currency, &rc.MachineRatePerHour, &rc.FilamentPricePerKg, &rc.FilamentKgPerHour,
		&rc.LaborFlat, &rc.FailureSurchargePct,
	)
	return rc, err
}

// UpsertRateCard updates costing defaults for a site.
func UpsertRateCard(ctx context.Context, pool *pgxpool.Pool, siteID string, rc api.RateCard) (api.RateCard, error) {
	if rc.Currency == "" {
		rc.Currency = "USD"
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO rate_cards (
			site_id, currency, machine_rate_per_hour, filament_price_per_kg, filament_kg_per_hour,
			labor_flat, failure_surcharge_pct, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7, now())
		ON CONFLICT (site_id) DO UPDATE SET
			currency=EXCLUDED.currency,
			machine_rate_per_hour=EXCLUDED.machine_rate_per_hour,
			filament_price_per_kg=EXCLUDED.filament_price_per_kg,
			filament_kg_per_hour=EXCLUDED.filament_kg_per_hour,
			labor_flat=EXCLUDED.labor_flat,
			failure_surcharge_pct=EXCLUDED.failure_surcharge_pct,
			updated_at=now()`,
		siteID, rc.Currency, rc.MachineRatePerHour, rc.FilamentPricePerKg, rc.FilamentKgPerHour,
		rc.LaborFlat, rc.FailureSurchargePct,
	)
	if err != nil {
		return api.RateCard{}, err
	}
	return GetRateCard(ctx, pool, siteID)
}
