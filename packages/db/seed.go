package db

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed seed/lab.sql
var labSeedSQL string

// SeedLab applies packages/db/seed/lab.sql (idempotent upserts for lab defaults).
func SeedLab(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("seed: nil pool")
	}
	if _, err := pool.Exec(ctx, labSeedSQL); err != nil {
		return fmt.Errorf("seed lab: %w", err)
	}
	return nil
}
