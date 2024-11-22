package server_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elleshadow/noPromises/pkg/server"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) *server.Server {
	// Create temporary test directories
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	webDir := filepath.Join(tmpDir, "web", "templates")
	migrationsDir := filepath.Join(tmpDir, "migrations")

	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "api"), 0755))
	require.NoError(t, os.MkdirAll(webDir, 0755))
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

	// Create test template
	require.NoError(t, os.WriteFile(
		filepath.Join(webDir, "index.html"),
		[]byte(`<!DOCTYPE html><html><body>{{.Title}}</body></html>`),
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
	srv, err := server.NewServer(server.Config{
		Port:           0, // Use random port
		DocsPath:       docsDir,
		DBPath:         ":memory:",    // Use in-memory SQLite for tests
		MigrationsPath: migrationsDir, // Use test migrations directory
	})
	require.NoError(t, err)
	require.NotNil(t, srv)

	return srv
}
