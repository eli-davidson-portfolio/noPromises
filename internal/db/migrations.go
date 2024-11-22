package db

import (
	"context"
	"fmt"
	"log"

	"github.com/elleshadow/noPromises/internal/db/migrations"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	db *DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *DB) *MigrationManager {
	return &MigrationManager{
		db: db,
	}
}

// ApplyMigrations applies all migrations from the given directory
func (m *MigrationManager) ApplyMigrations(ctx context.Context, migrationsDir string) error {
	log.Printf("[INFO] Starting database migrations from directory: %s", migrationsDir)

	migrator := migrations.NewMigrator(m.db.DB)

	migrations, err := migrations.LoadMigrations(migrationsDir)
	if err != nil {
		return fmt.Errorf("loading migrations: %w", err)
	}

	log.Printf("[INFO] Found %d migrations to apply", len(migrations))

	if err := migrator.ApplyMigrations(ctx, migrations); err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}

	log.Printf("[INFO] Successfully applied all migrations. Database schema is up to date")
	return nil
}

// RollbackMigration rolls back the last applied migration
func (m *MigrationManager) RollbackMigration(ctx context.Context) error {
	log.Println("[INFO] Rolling back last database migration")

	migrator := migrations.NewMigrator(m.db.DB)

	if err := migrator.RollbackMigration(ctx); err != nil {
		return fmt.Errorf("rolling back migration: %w", err)
	}

	log.Println("[INFO] Successfully rolled back last migration")
	return nil
}

// GetCurrentVersion returns the current schema version
func (m *MigrationManager) GetCurrentVersion(ctx context.Context) (int, error) {
	migrator := migrations.NewMigrator(m.db.DB)
	version, err := migrator.GetCurrentVersion(ctx)
	if err != nil {
		return 0, fmt.Errorf("getting schema version: %w", err)
	}
	log.Printf("[INFO] Current database schema version: %d", version)
	return version, nil
}
