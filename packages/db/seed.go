package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedLab applies LabSeedSQL (idempotent upserts for lab defaults).
func SeedLab(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("seed: nil pool")
	}
	if _, err := pool.Exec(ctx, LabSeedSQL); err != nil {
		return fmt.Errorf("seed lab: %w", err)
	}
	return nil
}
