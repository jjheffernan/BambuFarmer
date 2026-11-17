package db

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jjheffernan/BambuFarmer/packages/api"
)

func hasMaintenanceTag(tags []string) bool {
	for _, t := range tags {
		if strings.EqualFold(strings.TrimSpace(t), "maintenance") {
			return true
		}
	}
	return false
}

// ListPrinterSnapshots returns projected printers hydrated from last_snapshot
// (AMS, HMS, task, temps) with column values as the summary source of truth.
func ListPrinterSnapshots(ctx context.Context, pool *pgxpool.Pool, siteID string) ([]api.PrinterSnapshot, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, site_id, external_sn, name, model, online, gcode_state, mc_percent, mc_remaining_time,
		       COALESCE(maintenance, FALSE), COALESCE(last_snapshot::text, '')
		FROM printers WHERE site_id=$1 ORDER BY name, external_sn`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []api.PrinterSnapshot
	for rows.Next() {
		var id, site, sn, name, model, state, raw string
		var online, maint bool
		var pct, remain int
		if err := rows.Scan(&id, &site, &sn, &name, &model, &online, &state, &pct, &remain, &maint, &raw); err != nil {
			return nil, err
		}
		s := HydratePrinterSnapshot(id, site, sn, name, model, state, online, pct, remain, raw)
		s.Maintenance = maint || hasMaintenanceTag(s.Tags)
		out = append(out, s)
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

// SetPrinterMaintenance toggles the local maintenance flag (scheduler exclusion).
func SetPrinterMaintenance(ctx context.Context, pool *pgxpool.Pool, siteID, sn string, on bool) (api.PrinterSummary, error) {
	sn = strings.TrimSpace(sn)
	if sn == "" {
		return api.PrinterSummary{}, Validation("external_sn required")
	}
	tag, err := pool.Exec(ctx, `
		UPDATE printers SET maintenance=$3, updated_at=now()
		WHERE site_id=$1 AND external_sn=$2`, siteID, sn, on)
	if err != nil {
		return api.PrinterSummary{}, err
	}
	if tag.RowsAffected() == 0 {
		return api.PrinterSummary{}, api.NotFound("printer not found")
	}
	list, err := ListPrinters(ctx, pool, siteID)
	if err != nil {
		return api.PrinterSummary{}, err
	}
	for _, p := range list {
		if p.ExternalSN == sn {
			return p, nil
		}
	}
	return api.PrinterSummary{}, api.NotFound("printer not found")
}
