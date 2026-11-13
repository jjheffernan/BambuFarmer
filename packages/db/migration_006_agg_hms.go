package db

// Daily print-error / HMS code counts for analytics.
const migration006AggHMS = `
CREATE TABLE IF NOT EXISTS agg_hms_daily (
    day DATE NOT NULL,
    site_id TEXT NOT NULL,
    code TEXT NOT NULL,
    count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (day, site_id, code)
);
`
