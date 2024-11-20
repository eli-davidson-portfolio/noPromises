// internal/db/schema.go

package db

// Migration represents a database schema migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// migrations defines the database schema migrations
var migrations = []Migration{
	{
		Version:     1,
		Description: "Initial schema",
		SQL: `
			CREATE TABLE IF NOT EXISTS server_state (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL
			);
			CREATE TABLE IF NOT EXISTS users (
				id TEXT PRIMARY KEY,
				username TEXT UNIQUE NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS sessions (
				id TEXT PRIMARY KEY,
				user_id TEXT NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				expires_at TIMESTAMP NOT NULL,
				FOREIGN KEY (user_id) REFERENCES users(id)
			);
		`,
	},
	// Add more migrations here as needed
}

// Schema-related code here
