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
