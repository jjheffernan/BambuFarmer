-- Single-site MVP schema: sites, rate cards, versioned presets.

CREATE TABLE IF NOT EXISTS sites (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO sites (id, name)
VALUES ('default', 'Default Site')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS rate_cards (
    id BIGSERIAL PRIMARY KEY,
    site_id TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    currency TEXT NOT NULL DEFAULT 'USD',
    machine_rate_per_hour DOUBLE PRECISION NOT NULL DEFAULT 0,
    filament_price_per_kg DOUBLE PRECISION NOT NULL DEFAULT 0,
    filament_kg_per_hour DOUBLE PRECISION NOT NULL DEFAULT 0,
    labor_flat DOUBLE PRECISION NOT NULL DEFAULT 0,
    failure_surcharge_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (site_id)
);

INSERT INTO rate_cards (site_id)
VALUES ('default')
ON CONFLICT (site_id) DO NOTHING;

-- Current filament preset tip + immutable revision history.
CREATE TABLE IF NOT EXISTS filament_presets (
    id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    material TEXT NOT NULL DEFAULT '',
    vendor TEXT NOT NULL DEFAULT '',
    product_line TEXT NOT NULL DEFAULT '',
    color_hex TEXT NOT NULL DEFAULT '',
    density_g_cm3 DOUBLE PRECISION NOT NULL DEFAULT 0,
    diameter_mm DOUBLE PRECISION NOT NULL DEFAULT 1.75,
    nozzle_temp_min INT NOT NULL DEFAULT 0,
    nozzle_temp_max INT NOT NULL DEFAULT 0,
    bed_temp INT NOT NULL DEFAULT 0,
    tray_info_idx TEXT NOT NULL DEFAULT '',
    price_per_kg DOUBLE PRECISION NOT NULL DEFAULT 0,
    spoolman_filament_id BIGINT,
    notes TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS filament_preset_revisions (
    id BIGSERIAL PRIMARY KEY,
    preset_id TEXT NOT NULL REFERENCES filament_presets(id) ON DELETE CASCADE,
    version INT NOT NULL,
    snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (preset_id, version)
);

CREATE TABLE IF NOT EXISTS print_presets (
    id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    compatible_models JSONB NOT NULL DEFAULT '[]'::jsonb,
    nozzle_diameter_mm DOUBLE PRECISION NOT NULL DEFAULT 0.4,
    layer_height_mm DOUBLE PRECISION NOT NULL DEFAULT 0.2,
    infill_pct DOUBLE PRECISION NOT NULL DEFAULT 15,
    process_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    printer_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    default_for_model TEXT NOT NULL DEFAULT '',
    est_speed_factor DOUBLE PRECISION NOT NULL DEFAULT 1,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS print_preset_revisions (
    id BIGSERIAL PRIMARY KEY,
    preset_id TEXT NOT NULL REFERENCES print_presets(id) ON DELETE CASCADE,
    version INT NOT NULL,
    snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (preset_id, version)
);

CREATE INDEX IF NOT EXISTS idx_filament_presets_site ON filament_presets(site_id);
CREATE INDEX IF NOT EXISTS idx_print_presets_site ON print_presets(site_id);
