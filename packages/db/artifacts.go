package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// InsertArtifact records a staged .gcode.3mf (bytes live under ARTIFACTS_DIR/object_key).
func InsertArtifact(ctx context.Context, pool *pgxpool.Pool, siteID string, a api.Artifact) (api.Artifact, error) {
	id := a.ID
	if id == "" {
		id = newID()
	}
	key := a.ObjectKey
	if key == "" {
		key = id
	}
	name := a.Filename
	if name == "" {
		name = key
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO artifacts (id, site_id, filename, object_key, bytes, est_seconds, est_grams)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, siteID, name, key, a.Bytes, a.EstSeconds, a.EstGrams,
	)
	if err != nil {
		return api.Artifact{}, err
	}
	return GetArtifact(ctx, pool, siteID, id)
}

// GetArtifact loads one artifact metadata row.
func GetArtifact(ctx context.Context, pool *pgxpool.Pool, siteID, id string) (api.Artifact, error) {
	var a api.Artifact
	err := pool.QueryRow(ctx, `
		SELECT id, site_id, filename, object_key, bytes, est_seconds, est_grams, created_at
		FROM artifacts WHERE id=$1 AND site_id=$2`, id, siteID).Scan(
		&a.ID, &a.SiteID, &a.Filename, &a.ObjectKey, &a.Bytes, &a.EstSeconds, &a.EstGrams, &a.CreatedAt,
	)
	if err != nil {
		return a, MapQueryError(err)
	}
	return a, nil
}

// DeleteArtifact removes metadata for one artifact (jobs keep a null artifact_id).
func DeleteArtifact(ctx context.Context, pool *pgxpool.Pool, siteID, id string) error {
	tag, err := pool.Exec(ctx, `DELETE FROM artifacts WHERE id=$1 AND site_id=$2`, id, siteID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return api.NotFound("artifact")
	}
	return nil
}

// ListArtifacts returns recent artifacts for a site.
func ListArtifacts(ctx context.Context, pool *pgxpool.Pool, siteID string, limit int) ([]api.Artifact, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := pool.Query(ctx, `
		SELECT id, site_id, filename, object_key, bytes, est_seconds, est_grams, created_at
		FROM artifacts WHERE site_id=$1 ORDER BY created_at DESC LIMIT $2`, siteID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []api.Artifact
	for rows.Next() {
		var a api.Artifact
		var created time.Time
		if err := rows.Scan(&a.ID, &a.SiteID, &a.Filename, &a.ObjectKey, &a.Bytes, &a.EstSeconds, &a.EstGrams, &created); err != nil {
			return nil, err
		}
		a.CreatedAt = created
		out = append(out, a)
	}
	return out, rows.Err()
}
