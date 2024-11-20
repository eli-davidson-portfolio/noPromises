// internal/db/manager.go

package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SchemaManager struct {
	db *sql.DB
}

func NewSchemaManager(db *sql.DB) *SchemaManager {
	return &SchemaManager{db: db}
}

func (sm *SchemaManager) GetCurrentVersion(ctx context.Context) (int, error) {
	// Make sure schema_versions table exists
	_, err := sm.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_versions (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL,
			description TEXT
		);`)
	if err != nil {
		return 0, fmt.Errorf("creating schema_versions table: %w", err)
	}

	var version int
	err = sm.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) 
		FROM schema_versions`).Scan(&version)
	return version, err
}

func (sm *SchemaManager) Migrate(ctx context.Context) error {
	currentVersion, err := sm.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("getting current version: %w", err)
	}

	// Run each migration in a transaction
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		tx, err := sm.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("beginning transaction: %w", err)
		}

		// Run the migration
		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("rollback failed: %v (original error: %v)", rbErr, err)
			}
			return fmt.Errorf("executing migration %d: %w", migration.Version, err)
		}

		// Record the migration
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO schema_versions (version, applied_at, description)
			VALUES (?, ?, ?)`,
			migration.Version, time.Now(), migration.Description); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("rollback failed: %v (original error: %v)", rbErr, err)
			}
			return fmt.Errorf("recording migration %d: %w", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %d: %w", migration.Version, err)
		}
	}

	return nil
}
