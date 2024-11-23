-- migrations/000002_add_auth_tables.up.sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_login TIMESTAMP,
    is_admin BOOLEAN DEFAULT FALSE
);

CREATE TABLE token_blacklist (
    token TEXT PRIMARY KEY,
    revoked_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

-- Add indexes
CREATE INDEX idx_token_blacklist_expires ON token_blacklist(expires_at);