package db

import (
	"context"
	"strings"
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

// UpdateSite sets name and timezone for an existing site.
func UpdateSite(ctx context.Context, pool *pgxpool.Pool, id, name, timezone string) (api.Site, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	timezone = strings.TrimSpace(timezone)
	if id == "" {
		return api.Site{}, Validation("id required")
	}
	if name == "" {
		return api.Site{}, Validation("name required")
	}
	if timezone == "" {
		timezone = "UTC"
	}
	tag, err := pool.Exec(ctx, `
		UPDATE sites SET name=$1, timezone=$2 WHERE id=$3`, name, timezone, id)
	if err != nil {
		return api.Site{}, err
	}
	if tag.RowsAffected() == 0 {
		return api.Site{}, ErrNotFound
	}
	return GetSite(ctx, pool, id)
}
