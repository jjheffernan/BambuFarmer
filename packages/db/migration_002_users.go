package db

// Local operator accounts (Argon2id hashes).
const migration002Users = `
-- Local operator accounts (AUTH_MODE=lan). Passwords are Argon2id hashes only.

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    roles TEXT[] NOT NULL DEFAULT '{viewer}',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_username_unique UNIQUE (username)
);

CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower ON users (lower(username));
`
