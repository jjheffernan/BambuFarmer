package db

// LabSeedSQL is idempotent lab defaults (sites, rates, sample presets).
const LabSeedSQL = `
-- Lab seed: safe to re-run (ON CONFLICT). Assumes migrate/baseline already applied.
-- Applied by: make db-reset  ·  or db.SeedLab(ctx, pool)

INSERT INTO sites (id, name, timezone)
VALUES ('default', 'Default Site', 'UTC')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO rate_cards (
    site_id, currency, machine_rate_per_hour, filament_price_per_kg, filament_kg_per_hour
) VALUES ('default', 'USD', 5, 20, 0.1)
ON CONFLICT (site_id) DO UPDATE SET
    currency = EXCLUDED.currency,
    machine_rate_per_hour = EXCLUDED.machine_rate_per_hour,
    filament_price_per_kg = EXCLUDED.filament_price_per_kg,
    filament_kg_per_hour = EXCLUDED.filament_kg_per_hour,
    updated_at = now();

INSERT INTO filament_presets (
    id, site_id, name, version, material, vendor, color_hex, diameter_mm,
    nozzle_temp_min, nozzle_temp_max, bed_temp, price_per_kg, active
) VALUES (
    'pla-generic', 'default', 'Generic PLA', 1, 'PLA', 'Generic', '#00AE42', 1.75,
    190, 220, 55, 20, TRUE
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    material = EXCLUDED.material,
    vendor = EXCLUDED.vendor,
    color_hex = EXCLUDED.color_hex,
    diameter_mm = EXCLUDED.diameter_mm,
    nozzle_temp_min = EXCLUDED.nozzle_temp_min,
    nozzle_temp_max = EXCLUDED.nozzle_temp_max,
    bed_temp = EXCLUDED.bed_temp,
    price_per_kg = EXCLUDED.price_per_kg,
    active = EXCLUDED.active,
    updated_at = now();

INSERT INTO print_presets (
    id, site_id, name, version, compatible_models, nozzle_diameter_mm,
    layer_height_mm, infill_pct, est_speed_factor, active
) VALUES (
    '0.20-standard', 'default', '0.20mm Standard', 1,
    '["P1S","X1C","A1"]'::jsonb, 0.4, 0.2, 15, 1, TRUE
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    compatible_models = EXCLUDED.compatible_models,
    nozzle_diameter_mm = EXCLUDED.nozzle_diameter_mm,
    layer_height_mm = EXCLUDED.layer_height_mm,
    infill_pct = EXCLUDED.infill_pct,
    est_speed_factor = EXCLUDED.est_speed_factor,
    active = EXCLUDED.active,
    updated_at = now();

INSERT INTO install_state (id, completed, step, site_name, timezone, auth_mode)
VALUES (1, FALSE, 'welcome', 'Default Site', 'UTC', 'lan')
ON CONFLICT (id) DO NOTHING;
`
