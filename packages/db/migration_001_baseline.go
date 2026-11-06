package db

// Baseline schema (sites, presets, jobs, analytics, install_state).
const migration001Baseline = `
-- BambuFarmer schema baseline (squashed from former 001–004).
-- Applied as migration name 001_baseline.sql via db.Migrate.

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

-- Jobs, printer projections, artifacts (Phase 0/1 scaffold).

CREATE TABLE IF NOT EXISTS printers (
    id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    external_sn TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL DEFAULT '',
    online BOOLEAN NOT NULL DEFAULT FALSE,
    gcode_state TEXT NOT NULL DEFAULT 'IDLE',
    mc_percent INT NOT NULL DEFAULT 0,
    mc_remaining_time INT NOT NULL DEFAULT 0,
    last_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_seen_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (site_id, external_sn)
);

CREATE INDEX IF NOT EXISTS idx_printers_site ON printers(site_id);

CREATE TABLE IF NOT EXISTS artifacts (
    id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    object_key TEXT NOT NULL DEFAULT '',
    bytes BIGINT NOT NULL DEFAULT 0,
    est_seconds INT NOT NULL DEFAULT 0,
    est_grams DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'queued',
    printer_id TEXT REFERENCES printers(id) ON DELETE SET NULL,
    external_sn TEXT NOT NULL DEFAULT '',
    artifact_id TEXT REFERENCES artifacts(id) ON DELETE SET NULL,
    copies INT NOT NULL DEFAULT 1,
    est_seconds INT NOT NULL DEFAULT 0,
    est_grams DOUBLE PRECISION NOT NULL DEFAULT 0,
    cost_estimate DOUBLE PRECISION NOT NULL DEFAULT 0,
    cost_actual DOUBLE PRECISION NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_jobs_site_state ON jobs(site_id, state);

CREATE TABLE IF NOT EXISTS job_events (
    id BIGSERIAL PRIMARY KEY,
    job_id TEXT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    from_state TEXT NOT NULL DEFAULT '',
    to_state TEXT NOT NULL,
    actor TEXT NOT NULL DEFAULT 'system',
    reason TEXT NOT NULL DEFAULT '',
    at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_job_events_job ON job_events(job_id);

-- Analytics rollups (owned by analytics service; applied via shared migrate for MVP).

CREATE TABLE IF NOT EXISTS fact_job_completed (
    id BIGSERIAL PRIMARY KEY,
    job_id TEXT NOT NULL,
    site_id TEXT NOT NULL,
    printer_id TEXT NOT NULL DEFAULT '',
    external_sn TEXT NOT NULL DEFAULT '',
    duration_s INT NOT NULL DEFAULT 0,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    cost_actual DOUBLE PRECISION NOT NULL DEFAULT 0,
    grams DOUBLE PRECISION NOT NULL DEFAULT 0,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fact_job_completed_job ON fact_job_completed(job_id);

CREATE TABLE IF NOT EXISTS agg_utilization_daily (
    day DATE NOT NULL,
    site_id TEXT NOT NULL,
    printer_id TEXT NOT NULL DEFAULT '',
    external_sn TEXT NOT NULL DEFAULT '',
    idle_s INT NOT NULL DEFAULT 0,
    run_s INT NOT NULL DEFAULT 0,
    fail_s INT NOT NULL DEFAULT 0,
    samples INT NOT NULL DEFAULT 0,
    PRIMARY KEY (day, site_id, external_sn)
);

CREATE TABLE IF NOT EXISTS agg_cost_daily (
    day DATE NOT NULL,
    site_id TEXT NOT NULL,
    material_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    machine_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    labor_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    jobs INT NOT NULL DEFAULT 0,
    PRIMARY KEY (day, site_id)
);

CREATE TABLE IF NOT EXISTS analytics_counters (
    key TEXT PRIMARY KEY,
    value BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- First-run install wizard state (single-row).

CREATE TABLE IF NOT EXISTS install_state (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    version TEXT NOT NULL DEFAULT '',
    step TEXT NOT NULL DEFAULT 'welcome',
    site_name TEXT NOT NULL DEFAULT 'Default Site',
    timezone TEXT NOT NULL DEFAULT 'UTC',
    auth_mode TEXT NOT NULL DEFAULT 'lan',
    farm_manager_url TEXT NOT NULL DEFAULT '',
    farm_manager_transport TEXT NOT NULL DEFAULT 'fm_native',
    spoolman_url TEXT NOT NULL DEFAULT '',
    manyfold_url TEXT NOT NULL DEFAULT '',
    admin_username TEXT NOT NULL DEFAULT 'admin',
    notes TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO install_state (id)
VALUES (1)
ON CONFLICT (id) DO NOTHING;
`
