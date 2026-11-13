package db

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

// ListPrinterSnapshots returns projected printers hydrated from last_snapshot
// (AMS, HMS, task, temps) with column values as the summary source of truth.
func ListPrinterSnapshots(ctx context.Context, pool *pgxpool.Pool, siteID string) ([]api.PrinterSnapshot, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, site_id, external_sn, name, model, online, gcode_state, mc_percent, mc_remaining_time, COALESCE(last_snapshot::text, '')
		FROM printers WHERE site_id=$1 ORDER BY name, external_sn`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []api.PrinterSnapshot
	for rows.Next() {
		var id, site, sn, name, model, state, raw string
		var online bool
		var pct, remain int
		if err := rows.Scan(&id, &site, &sn, &name, &model, &online, &state, &pct, &remain, &raw); err != nil {
			return nil, err
		}
		out = append(out, HydratePrinterSnapshot(id, site, sn, name, model, state, online, pct, remain, raw))
	}
	return out, rows.Err()
}

// HydratePrinterSnapshot merges projection columns with a last_snapshot JSON payload.
func HydratePrinterSnapshot(id, siteID, sn, name, model, state string, online bool, pct, remain int, raw string) api.PrinterSnapshot {
	var s api.PrinterSnapshot
	if raw != "" && raw != "null" {
		_ = json.Unmarshal([]byte(raw), &s)
	}
	s.ID = id
	s.SiteID = siteID
	s.ExternalSN = sn
	if name != "" {
		s.Name = name
	}
	if model != "" {
		s.Model = model
	}
	s.Online = online
	if state != "" {
		s.GcodeState = state
	}
	s.MCPercent = pct
	s.RemainingS = remain
	if len(s.Tags) == 0 {
		s.Tags = tagsFromSnapshot(raw)
	}
	return s
}
