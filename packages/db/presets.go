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

// ListFilamentPresets returns current tips for a site.
func ListFilamentPresets(ctx context.Context, pool *pgxpool.Pool, siteID string) ([]api.FilamentPreset, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, name, version, material, vendor, product_line, color_hex,
		       density_g_cm3, diameter_mm, nozzle_temp_min, nozzle_temp_max, bed_temp,
		       tray_info_idx, price_per_kg, spoolman_filament_id, notes, active, updated_at
		FROM filament_presets
		WHERE site_id=$1
		ORDER BY name`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []api.FilamentPreset
	for rows.Next() {
		var p api.FilamentPreset
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Version, &p.Material, &p.Vendor, &p.ProductLine, &p.ColorHex,
			&p.DensityGCm3, &p.DiameterMM, &p.NozzleTempMin, &p.NozzleTempMax, &p.BedTemp,
			&p.TrayInfoIdx, &p.PricePerKg, &p.SpoolmanFilamentID, &p.Notes, &p.Active, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpsertFilamentPreset inserts or updates a tip and appends a revision snapshot.
func UpsertFilamentPreset(ctx context.Context, pool *pgxpool.Pool, siteID string, p api.FilamentPreset) (api.FilamentPreset, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return p, err
	}
	defer tx.Rollback(ctx)

	var existingVersion int
	err = tx.QueryRow(ctx, `SELECT version FROM filament_presets WHERE id=$1 FOR UPDATE`, p.ID).Scan(&existingVersion)
	if err != nil && err != pgx.ErrNoRows {
		return p, err
	}
	if err == pgx.ErrNoRows {
		p.Version = 1
	} else {
		p.Version = existingVersion + 1
	}
	p.UpdatedAt = time.Now().UTC()
	if p.DiameterMM == 0 {
		p.DiameterMM = 1.75
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO filament_presets (
			id, site_id, name, version, material, vendor, product_line, color_hex,
			density_g_cm3, diameter_mm, nozzle_temp_min, nozzle_temp_max, bed_temp,
			tray_info_idx, price_per_kg, spoolman_filament_id, notes, active, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,
			$9,$10,$11,$12,$13,
			$14,$15,$16,$17,$18,$19
		)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, version=EXCLUDED.version, material=EXCLUDED.material,
			vendor=EXCLUDED.vendor, product_line=EXCLUDED.product_line, color_hex=EXCLUDED.color_hex,
			density_g_cm3=EXCLUDED.density_g_cm3, diameter_mm=EXCLUDED.diameter_mm,
			nozzle_temp_min=EXCLUDED.nozzle_temp_min, nozzle_temp_max=EXCLUDED.nozzle_temp_max,
			bed_temp=EXCLUDED.bed_temp, tray_info_idx=EXCLUDED.tray_info_idx,
			price_per_kg=EXCLUDED.price_per_kg, spoolman_filament_id=EXCLUDED.spoolman_filament_id,
			notes=EXCLUDED.notes, active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		p.ID, siteID, p.Name, p.Version, p.Material, p.Vendor, p.ProductLine, p.ColorHex,
		p.DensityGCm3, p.DiameterMM, p.NozzleTempMin, p.NozzleTempMax, p.BedTemp,
		p.TrayInfoIdx, p.PricePerKg, p.SpoolmanFilamentID, p.Notes, p.Active, p.UpdatedAt,
	)
	if err != nil {
		return p, err
	}

	snap, err := json.Marshal(p)
	if err != nil {
		return p, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO filament_preset_revisions (preset_id, version, snapshot)
		VALUES ($1,$2,$3)`, p.ID, p.Version, snap); err != nil {
		return p, fmt.Errorf("revision: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return p, err
	}
	return p, nil
}

// ListPrintPresets returns current tips for a site.
func ListPrintPresets(ctx context.Context, pool *pgxpool.Pool, siteID string) ([]api.PrintPreset, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, name, version, compatible_models, nozzle_diameter_mm, layer_height_mm, infill_pct,
		       process_json, printer_json, default_for_model, est_speed_factor, active, updated_at
		FROM print_presets
		WHERE site_id=$1
		ORDER BY name`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []api.PrintPreset
	for rows.Next() {
		var p api.PrintPreset
		var models, process, printer []byte
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Version, &models, &p.NozzleDiameterMM, &p.LayerHeightMM, &p.InfillPct,
			&process, &printer, &p.DefaultForModel, &p.EstSpeedFactor, &p.Active, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(models, &p.CompatibleModels)
		_ = json.Unmarshal(process, &p.ProcessJSON)
		_ = json.Unmarshal(printer, &p.PrinterJSON)
		if p.CompatibleModels == nil {
			p.CompatibleModels = []string{}
		}
		if p.ProcessJSON == nil {
			p.ProcessJSON = map[string]any{}
		}
		if p.PrinterJSON == nil {
			p.PrinterJSON = map[string]any{}
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpsertPrintPreset inserts or updates a tip and appends a revision snapshot.
func UpsertPrintPreset(ctx context.Context, pool *pgxpool.Pool, siteID string, p api.PrintPreset) (api.PrintPreset, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return p, err
	}
	defer tx.Rollback(ctx)

	var existingVersion int
	err = tx.QueryRow(ctx, `SELECT version FROM print_presets WHERE id=$1 FOR UPDATE`, p.ID).Scan(&existingVersion)
	if err != nil && err != pgx.ErrNoRows {
		return p, err
	}
	if err == pgx.ErrNoRows {
		p.Version = 1
	} else {
		p.Version = existingVersion + 1
	}
	p.UpdatedAt = time.Now().UTC()
	if p.CompatibleModels == nil {
		p.CompatibleModels = []string{}
	}
	if p.ProcessJSON == nil {
		p.ProcessJSON = map[string]any{}
	}
	if p.PrinterJSON == nil {
		p.PrinterJSON = map[string]any{}
	}
	if p.EstSpeedFactor == 0 {
		p.EstSpeedFactor = 1
	}

	models, _ := json.Marshal(p.CompatibleModels)
	process, _ := json.Marshal(p.ProcessJSON)
	printer, _ := json.Marshal(p.PrinterJSON)

	_, err = tx.Exec(ctx, `
		INSERT INTO print_presets (
			id, site_id, name, version, compatible_models, nozzle_diameter_mm, layer_height_mm, infill_pct,
			process_json, printer_json, default_for_model, est_speed_factor, active, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,
			$9,$10,$11,$12,$13,$14
		)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, version=EXCLUDED.version, compatible_models=EXCLUDED.compatible_models,
			nozzle_diameter_mm=EXCLUDED.nozzle_diameter_mm, layer_height_mm=EXCLUDED.layer_height_mm,
			infill_pct=EXCLUDED.infill_pct, process_json=EXCLUDED.process_json, printer_json=EXCLUDED.printer_json,
			default_for_model=EXCLUDED.default_for_model, est_speed_factor=EXCLUDED.est_speed_factor,
			active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		p.ID, siteID, p.Name, p.Version, models, p.NozzleDiameterMM, p.LayerHeightMM, p.InfillPct,
		process, printer, p.DefaultForModel, p.EstSpeedFactor, p.Active, p.UpdatedAt,
	)
	if err != nil {
		return p, err
	}

	snap, err := json.Marshal(p)
	if err != nil {
		return p, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO print_preset_revisions (preset_id, version, snapshot)
		VALUES ($1,$2,$3)`, p.ID, p.Version, snap); err != nil {
		return p, fmt.Errorf("revision: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return p, err
	}
	return p, nil
}

// ImportPresetBundle replaces nothing globally; upserts each preset (version bumping).
func ImportPresetBundle(ctx context.Context, pool *pgxpool.Pool, siteID string, bundle api.PresetBundle) error {
	for _, f := range bundle.FilamentPresets {
		if f.ID == "" {
			return fmt.Errorf("filament preset missing id")
		}
		if _, err := UpsertFilamentPreset(ctx, pool, siteID, f); err != nil {
			return err
		}
	}
	for _, p := range bundle.PrintPresets {
		if p.ID == "" {
			return fmt.Errorf("print preset missing id")
		}
		if _, err := UpsertPrintPreset(ctx, pool, siteID, p); err != nil {
			return err
		}
	}
	return nil
}

// ExportPresetBundle returns current tips.
func ExportPresetBundle(ctx context.Context, pool *pgxpool.Pool, siteID string) (api.PresetBundle, error) {
	fil, err := ListFilamentPresets(ctx, pool, siteID)
	if err != nil {
		return api.PresetBundle{}, err
	}
	pr, err := ListPrintPresets(ctx, pool, siteID)
	if err != nil {
		return api.PresetBundle{}, err
	}
	if fil == nil {
		fil = []api.FilamentPreset{}
	}
	if pr == nil {
		pr = []api.PrintPreset{}
	}
	return api.PresetBundle{FilamentPresets: fil, PrintPresets: pr}, nil
}
