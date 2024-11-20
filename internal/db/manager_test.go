// internal/db/manager_test.go

package db

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver

	"github.com/stretchr/testify/require"
)

func TestSchemaManager(t *testing.T) {
	// Create in-memory database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	manager := NewSchemaManager(db)
	ctx := context.Background()

	// Test initial version
	version, err := manager.GetCurrentVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, version)

	// Test migrations
	err = manager.Migrate(ctx)
	require.NoError(t, err)

	// Verify final version
	version, err = manager.GetCurrentVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, len(migrations), version)

	// Verify tables exist
	tables := []string{"server_state", "users", "sessions"}
	for _, table := range tables {
		var count int
		err := db.QueryRowContext(ctx, `
            SELECT COUNT(*) FROM sqlite_master 
            WHERE type='table' AND name=?`, table).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	}
}
