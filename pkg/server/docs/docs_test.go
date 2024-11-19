package docs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsServer(t *testing.T) {
	// Create test docs directory
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "api"), 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "test.md"),
		[]byte("# Test Documentation"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "api", "swagger.json"),
		[]byte(`{"openapi":"3.0.0"}`),
		0644,
	))

	// Create docs server
	srv := NewServer(Config{
		DocsPath: tmpDir,
	})
	require.NotNil(t, srv)

	// Setup routes
	srv.SetupRoutes()

	tests := []struct {
		name           string
		path           string
		setup          func(*Server)
		expectedStatus int
		expectedType   string
		expectedBody   string
	}{
		{
			name: "serve static documentation",
			path: "/docs/test.md",
			setup: func(_ *Server) {
				// No setup needed for static files
			},
			expectedStatus: http.StatusOK,
			expectedType:   "text/markdown; charset=utf-8",
			expectedBody:   "# Test Documentation",
		},
		{
			name: "serve swagger json",
			path: "/api/swagger.json",
			setup: func(_ *Server) {
				// No setup needed for static files
			},
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
			expectedBody:   `{"openapi":"3.0.0"}`,
		},
		{
			name: "generate network diagram",
			path: "/diagrams/network/test-flow",
			setup: func(s *Server) {
				s.mermaidGen.SetNetwork("test-flow", map[string]interface{}{
					"nodes": map[string]interface{}{
						"reader": map[string]interface{}{
							"type":   "FileReader",
							"status": "running",
						},
						"writer": map[string]interface{}{
							"type":   "FileWriter",
							"status": "waiting",
						},
					},
					"edges": []interface{}{
						map[string]interface{}{
							"from": "reader",
							"to":   "writer",
							"port": "data",
						},
					},
				})
			},
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
		{
			name: "serve swagger UI",
			path: "/api-docs",
			setup: func(_ *Server) {
				// No setup needed for Swagger UI
			},
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(srv)
			}

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			srv.Router().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), tt.expectedType)

			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, strings.TrimSpace(w.Body.String()))
			}

			if tt.path == "/diagrams/network/test-flow" {
				var response struct {
					Diagram string `json:"diagram"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Contains(t, response.Diagram, "reader[FileReader]:::running")
				assert.Contains(t, response.Diagram, "writer[FileWriter]:::waiting")
				assert.Contains(t, response.Diagram, "reader -->|data| writer")
			}
		})
	}
}

func TestLiveUpdates(t *testing.T) {
	srv := NewServer(Config{
		DocsPath: "testdata/docs",
	})
	srv.SetupRoutes()

	req := httptest.NewRequest("GET", "/diagrams/network/test-flow/live", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusSwitchingProtocols, w.Code)
}
