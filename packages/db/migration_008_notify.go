package db

const migration008Notify = `
CREATE TABLE IF NOT EXISTS notify_channels (
    id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    event_types JSONB NOT NULL DEFAULT '["job.finished","job.failed","alert.raised"]'::jsonb,
    auth_header TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notify_channels_site_idx ON notify_channels (site_id);

CREATE TABLE IF NOT EXISTS notify_printers (
    external_sn TEXT PRIMARY KEY,
    site_id TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    channel_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notify_printers_site_idx ON notify_printers (site_id);
`