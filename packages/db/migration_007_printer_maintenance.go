package db

const migration007PrinterMaintenance = `
ALTER TABLE printers ADD COLUMN IF NOT EXISTS maintenance BOOLEAN NOT NULL DEFAULT FALSE;
`