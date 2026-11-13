package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// PresetRevision is one immutable history row.
type PresetRevision struct {
	Version   int             `json:"version"`
	CreatedAt time.Time       `json:"created_at"`
	Snapshot  json.RawMessage `json:"snapshot"`
}

// ListFilamentPresetRevisions returns history newest-first for a site-owned preset.
func ListFilamentPresetRevisions(ctx context.Context, pool *pgxpool.Pool, siteID, presetID string) ([]PresetRevision, error) {
	if err := ensureFilamentPreset(ctx, pool, siteID, presetID); err != nil {
		return nil, err
	}
	return listRevisions(ctx, pool, `filament_preset_revisions`, presetID)
}

// ListPrintPresetRevisions returns history newest-first for a site-owned preset.
func ListPrintPresetRevisions(ctx context.Context, pool *pgxpool.Pool, siteID, presetID string) ([]PresetRevision, error) {
	if err := ensurePrintPreset(ctx, pool, siteID, presetID); err != nil {
		return nil, err
	}
	return listRevisions(ctx, pool, `print_preset_revisions`, presetID)
}

// RestoreFilamentPresetRevision loads a revision snapshot and upserts it as a new tip version.
func RestoreFilamentPresetRevision(ctx context.Context, pool *pgxpool.Pool, siteID, presetID string, version int) (api.FilamentPreset, error) {
	if version < 1 {
		return api.FilamentPreset{}, Validation("version required")
	}
	raw, err := loadRevision(ctx, pool, siteID, presetID, version, `filament_presets`, `filament_preset_revisions`)
	if err != nil {
		return api.FilamentPreset{}, err
	}
	var p api.FilamentPreset
	if err := json.Unmarshal(raw, &p); err != nil {
		return api.FilamentPreset{}, fmt.Errorf("revision snapshot: %w", err)
	}
	p.ID = presetID
	return UpsertFilamentPreset(ctx, pool, siteID, p)
}

// RestorePrintPresetRevision loads a revision snapshot and upserts it as a new tip version.
func RestorePrintPresetRevision(ctx context.Context, pool *pgxpool.Pool, siteID, presetID string, version int) (api.PrintPreset, error) {
	if version < 1 {
		return api.PrintPreset{}, Validation("version required")
	}
	raw, err := loadRevision(ctx, pool, siteID, presetID, version, `print_presets`, `print_preset_revisions`)
	if err != nil {
		return api.PrintPreset{}, err
	}
	var p api.PrintPreset
	if err := json.Unmarshal(raw, &p); err != nil {
		return api.PrintPreset{}, fmt.Errorf("revision snapshot: %w", err)
	}
	p.ID = presetID
	return UpsertPrintPreset(ctx, pool, siteID, p)
}

func listRevisions(ctx context.Context, pool *pgxpool.Pool, table, presetID string) ([]PresetRevision, error) {
	q := fmt.Sprintf(`SELECT version, created_at, snapshot FROM %s WHERE preset_id=$1 ORDER BY version DESC`, table)
	rows, err := pool.Query(ctx, q, presetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PresetRevision
	for rows.Next() {
		var r PresetRevision
		if err := rows.Scan(&r.Version, &r.CreatedAt, &r.Snapshot); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func ensureFilamentPreset(ctx context.Context, pool *pgxpool.Pool, siteID, presetID string) error {
	var id string
	err := pool.QueryRow(ctx, `SELECT id FROM filament_presets WHERE id=$1 AND site_id=$2`, presetID, siteID).Scan(&id)
	return MapQueryError(err)
}

func ensurePrintPreset(ctx context.Context, pool *pgxpool.Pool, siteID, presetID string) error {
	var id string
	err := pool.QueryRow(ctx, `SELECT id FROM print_presets WHERE id=$1 AND site_id=$2`, presetID, siteID).Scan(&id)
	return MapQueryError(err)
}

func loadRevision(ctx context.Context, pool *pgxpool.Pool, siteID, presetID string, version int, tipTable, revTable string) ([]byte, error) {
	var tip string
	qTip := fmt.Sprintf(`SELECT id FROM %s WHERE id=$1 AND site_id=$2`, tipTable)
	if err := MapQueryError(pool.QueryRow(ctx, qTip, presetID, siteID).Scan(&tip)); err != nil {
		return nil, err
	}
	var raw []byte
	qRev := fmt.Sprintf(`SELECT snapshot FROM %s WHERE preset_id=$1 AND version=$2`, revTable)
	err := pool.QueryRow(ctx, qRev, presetID, version).Scan(&raw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, MapQueryError(err)
		}
		return nil, err
	}
	return raw, nil
}
