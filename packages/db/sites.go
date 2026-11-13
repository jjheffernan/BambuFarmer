package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// ListSites returns all sites (single-site MVP usually one row).
func ListSites(ctx context.Context, pool *pgxpool.Pool) ([]api.Site, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, name, timezone, created_at FROM sites ORDER BY name, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []api.Site{}
	for rows.Next() {
		var s api.Site
		var created time.Time
		if err := rows.Scan(&s.ID, &s.Name, &s.Timezone, &created); err != nil {
			return nil, err
		}
		s.CreatedAt = created
		out = append(out, s)
	}
	return out, rows.Err()
}

// GetSite loads one site by id.
func GetSite(ctx context.Context, pool *pgxpool.Pool, id string) (api.Site, error) {
	var s api.Site
	err := pool.QueryRow(ctx, `
		SELECT id, name, timezone, created_at FROM sites WHERE id=$1`, id).Scan(
		&s.ID, &s.Name, &s.Timezone, &s.CreatedAt)
	if err != nil {
		return s, MapQueryError(err)
	}
	return s, nil
}
