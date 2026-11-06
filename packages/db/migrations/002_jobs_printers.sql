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
