package migrations

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestMigrations(t *testing.T) {
	db := setupTestDB(t)
	migrator := NewMigrator(db)

	// Create test migrations directory
	tmpDir := t.TempDir()

	// Create test migration files
	upSQL := "CREATE TABLE test (id INTEGER PRIMARY KEY);"
	downSQL := "DROP TABLE IF EXISTS test;"

	err := os.WriteFile(
		filepath.Join(tmpDir, "000001_create_test.up.sql"),
		[]byte(upSQL),
		0644,
	)
	require.NoError(t, err)

	err = os.WriteFile(
		filepath.Join(tmpDir, "000001_create_test.down.sql"),
		[]byte(downSQL),
		0644,
	)
	require.NoError(t, err)

	t.Run("load migrations", func(t *testing.T) {
		migrations, err := LoadMigrations(tmpDir)
		require.NoError(t, err)
		require.Len(t, migrations, 1, "Expected exactly one migration")

		migration := migrations[0]
		assert.Equal(t, 1, migration.Version)
		assert.Equal(t, "create_test", migration.Name)
		assert.Equal(t, upSQL, migration.UpScript)
		assert.Equal(t, downSQL, migration.DownScript)
	})

	t.Run("apply migrations", func(t *testing.T) {
		// First create migrations table
		err := migrator.createMigrationsTable(context.Background())
		require.NoError(t, err)

		// Load and apply migrations
		migrations, err := LoadMigrations(tmpDir)
		require.NoError(t, err)

		err = migrator.ApplyMigrations(context.Background(), migrations)
		require.NoError(t, err)

		// Verify table was created
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test'").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify migration was recorded
		var version int
		err = db.QueryRow("SELECT version FROM schema_version WHERE name = 'create_test'").Scan(&version)
		require.NoError(t, err)
		assert.Equal(t, 1, version)
	})

	t.Run("rollback migration", func(t *testing.T) {
		err := migrator.RollbackMigration(context.Background())
		require.NoError(t, err)

		// Verify table was dropped
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test'").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Verify migration was removed
		count = 0
		err = db.QueryRow("SELECT COUNT(*) FROM schema_version WHERE name = 'create_test'").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
