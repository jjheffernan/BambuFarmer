package db

// Optional FM file id on jobs (distinct from our artifact_id / object key).
const migration003JobFile = `
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS external_file_id TEXT NOT NULL DEFAULT '';
`
