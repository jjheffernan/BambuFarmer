package db

// Queue order, model target, and wait-on-previous for retained jobs.
const migration004JobQueue = `
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS queue_pos INT NOT NULL DEFAULT 0;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS target_model TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS after_job_id TEXT NOT NULL DEFAULT '';
`
