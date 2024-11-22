package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// Migration represents a database migration
type Migration struct {
	Version    int
	Name       string
	UpScript   string
	DownScript string
}

// Migrator handles database migrations
type Migrator struct {
	db *sql.DB
}

// NewMigrator creates a new migrator
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// ApplyMigrations applies the given migrations
func (m *Migrator) ApplyMigrations(ctx context.Context, migrations []Migration) error {
	if err := m.createMigrationsTable(ctx); err != nil {
		return err
	}

	for _, migration := range migrations {
		// Check if migration has already been applied
		var exists bool
		err := m.db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM schema_version WHERE version = ?
			)
		`, migration.Version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("checking migration status: %w", err)
		}

		if exists {
			log.Printf("[INFO] Skipping migration %d: %s (already applied)", migration.Version, migration.Name)
			continue
		}

		log.Printf("[INFO] Applying migration %d: %s", migration.Version, migration.Name)
		if err := m.applyMigration(ctx, migration); err != nil {
			return err
		}
		log.Printf("[INFO] Successfully applied migration %d", migration.Version)
	}

	return nil
}

// RollbackMigration rolls back the last applied migration
func (m *Migrator) RollbackMigration(ctx context.Context) error {
	var migration Migration
	err := m.db.QueryRowContext(ctx, `
		SELECT version, name, up_script, down_script 
		FROM schema_version 
		ORDER BY version DESC 
		LIMIT 1
	`).Scan(&migration.Version, &migration.Name, &migration.UpScript, &migration.DownScript)
	if err != nil {
		return fmt.Errorf("getting last migration: %w", err)
	}

	log.Printf("[INFO] Rolling back migration %d: %s", migration.Version, migration.Name)

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("[WARN] Error rolling back transaction: %v", err)
		}
	}()

	if _, err := tx.ExecContext(ctx, migration.DownScript); err != nil {
		return fmt.Errorf("executing down script: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM schema_version WHERE version = ?
	`, migration.Version); err != nil {
		return fmt.Errorf("removing migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	log.Printf("[INFO] Successfully rolled back migration %d", migration.Version)
	return nil
}

func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			up_script TEXT NOT NULL,
			down_script TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// applyMigration applies a single migration within a transaction
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("[WARN] Error rolling back transaction: %v", err)
		}
	}()

	// Execute the migration
	if _, err := tx.ExecContext(ctx, migration.UpScript); err != nil {
		return fmt.Errorf("executing up script: %w", err)
	}

	// Record the migration
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO schema_version (version, name, up_script, down_script)
		VALUES (?, ?, ?, ?)
	`, migration.Version, migration.Name, migration.UpScript, migration.DownScript); err != nil {
		return fmt.Errorf("recording migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// GetCurrentVersion returns the current schema version
func (m *Migrator) GetCurrentVersion(ctx context.Context) (int, error) {
	var version int
	err := m.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0)
		FROM schema_version
	`).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("getting current version: %w", err)
	}
	return version, nil
}
