-- Migration: add auth tables
-- Down migration

DROP INDEX IF EXISTS idx_token_blacklist_expires;
DROP TABLE IF EXISTS token_blacklist;
DROP TABLE IF EXISTS users;
