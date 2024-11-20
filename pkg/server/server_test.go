package server

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*Server, string) {
	// Create temporary test directories
	testDir := t.TempDir()
	docsDir := filepath.Join(testDir, "docs")

	// Create required directories
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "api"), 0755))

	// Create required files
	require.NoError(t, os.WriteFile(
		filepath.Join(docsDir, "README.md"),
		[]byte("# Test Documentation"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(docsDir, "api", "swagger.json"),
		[]byte(`{"openapi":"3.0.0"}`),
		0644,
	))

	// Create web templates directory in the project root
	webTemplatesDir := "web/templates"
	require.NoError(t, os.MkdirAll(webTemplatesDir, 0755))

	// Create template file
	require.NoError(t, os.WriteFile(
		filepath.Join(webTemplatesDir, "index.html"),
		[]byte(`<!DOCTYPE html><html><body>{{.Title}}</body></html>`),
		0644,
	))

	// Cleanup function to remove the web templates after test
	t.Cleanup(func() {
		os.RemoveAll("web")
	})

	// Create server with test configuration using port 0 (random available port)
	srv, err := NewServer(Config{
		Port:     0,
		DocsPath: docsDir,
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
	assert.NotNil(t, srv.webServer)
}

func TestServerConfiguration(t *testing.T) {
	t.Run("invalid docs path", func(t *testing.T) {
		srv, err := NewServer(Config{
			Port:     8080,
			DocsPath: "/invalid/path",
		})
		assert.Error(t, err)
		assert.Nil(t, srv)
	})

	t.Run("missing required files", func(t *testing.T) {
		tmpDir := t.TempDir()
		srv, err := NewServer(Config{
			Port:     8080,
			DocsPath: tmpDir,
		})
		assert.Error(t, err)
		assert.Nil(t, srv)
	})
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

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for server shutdown")
	}
}
