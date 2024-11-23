package docs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentationSystem(t *testing.T) {
	// Create a real docs structure
	tmpDir := t.TempDir()

	// Create directory structure
	dirs := []string{
		"api",
		"guides",
		"architecture",
		"architecture/patterns",
		"static/css", // Add static files directory
	}

	for _, dir := range dirs {
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, dir), 0755))
	}

	// Create test documentation files
	files := map[string]string{
		"README.md": "# Main Documentation\n\nWelcome to the documentation.",
		"guides/getting-started.md": `# Getting Started
## Installation
1. First step
2. Second step`,
		"api/swagger.json":               `{"openapi":"3.0.0"}`,
		"static/css/github-markdown.css": `.markdown-body { font-family: sans-serif; }`,
	}

	for path, content := range files {
		require.NoError(t, os.WriteFile(
			filepath.Join(tmpDir, path),
			[]byte(content),
			0644,
		))
	}

	// Create server with real docs
	srv := NewServer(Config{
		DocsPath: tmpDir,
	})
	require.NotNil(t, srv)

	// Setup routes once
	srv.SetupRoutes()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedType   string
		contains       []string
	}{
		{
			name:           "main readme",
			path:           "/docs/README.md",
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			contains:       []string{"Main Documentation", "Welcome to the documentation"},
		},
		{
			name:           "getting started",
			path:           "/docs/guides/getting-started.md",
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			contains:       []string{"Getting Started", "Installation"},
		},
		{
			name:           "swagger json",
			path:           "/docs/api/swagger.json",
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
			contains:       []string{`"openapi":"3.0.0"`},
		},
		{
			name:           "static css",
			path:           "/static/css/github-markdown.css",
			expectedStatus: http.StatusOK,
			expectedType:   "text/css; charset=utf-8",
			contains:       []string{".markdown-body"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
			assert.Equal(t, tt.expectedType, w.Header().Get("Content-Type"), "Content type mismatch")

			for _, s := range tt.contains {
				assert.Contains(t, w.Body.String(), s, "Response missing expected content")
			}

			// Add debug logging for failures
			if t.Failed() {
				t.Logf("Response body: %s", w.Body.String())
				t.Logf("Response headers: %v", w.Header())
			}
		})
	}
}
