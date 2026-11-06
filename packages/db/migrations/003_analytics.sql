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
