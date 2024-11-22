package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*Server, string) {
	// Create temporary test directories
	testDir := t.TempDir()
	docsDir := filepath.Join(testDir, "docs")
	migrationsDir := filepath.Join(testDir, "migrations")
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "api"), 0755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create required files
	require.NoError(t, os.WriteFile(
		filepath.Join(docsDir, "README.md"),
		[]byte("# Test Documentation"),
		0644,
	))

	require.NoError(t, os.WriteFile(
		filepath.Join(docsDir, "api", "swagger.json"),
		[]byte("{}"),
		0644,
	))

	// Create test migrations
	require.NoError(t, os.WriteFile(
		filepath.Join(migrationsDir, "000001_initial_schema.up.sql"),
		[]byte(`
			CREATE TABLE IF NOT EXISTS flows (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				config JSON NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
		`),
		0644,
	))

	require.NoError(t, os.WriteFile(
		filepath.Join(migrationsDir, "000001_initial_schema.down.sql"),
		[]byte("DROP TABLE IF EXISTS flows;"),
		0644,
	))

	// Create server with test configuration
	srv, err := NewServer(Config{
		Port:           0,
		DocsPath:       docsDir,
		DBPath:         ":memory:",    // Use in-memory SQLite for tests
		MigrationsPath: migrationsDir, // Use test migrations directory
	})
	require.NoError(t, err)
	require.NotNil(t, srv)

	return srv, testDir
}

func TestNewServer(t *testing.T) {
	srv, _ := setupTestServer(t)
	assert.NotNil(t, srv)
	assert.NotNil(t, srv.router)
	assert.NotNil(t, srv.flows)
	assert.NotNil(t, srv.processes)
}

func TestServerLifecycle(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Test server start/stop
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to trigger shutdown
	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}
