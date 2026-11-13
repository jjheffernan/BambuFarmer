package db

// Optional Spoolman spool id used when a job finishes.
const migration005JobSpool = `
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS spool_id TEXT NOT NULL DEFAULT '';
`
