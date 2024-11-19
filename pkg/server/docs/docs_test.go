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
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "serve static documentation",
			path: "/test.md",
			setup: func(_ *Server) {
				// No setup needed for static files
			},
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				bodyStr := w.Body.String()
				assert.Contains(t, bodyStr, "<html>")
				assert.Contains(t, bodyStr, "<div class=\"markdown-body\">")
				assert.Contains(t, bodyStr, "# Test Documentation")
			},
		},
		{
			name: "serve swagger json",
			path: "/api/swagger.json",
			setup: func(_ *Server) {
				// No setup needed for static files
			},
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, `{"openapi":"3.0.0"}`, strings.TrimSpace(w.Body.String()))
			},
		},
		{
			name: "serve swagger UI",
			path: "/api-docs",
			setup: func(_ *Server) {
				// No setup needed for Swagger UI
			},
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				bodyStr := w.Body.String()
				assert.Contains(t, bodyStr, "<html>")
				assert.Contains(t, bodyStr, "swagger-ui")
				assert.Contains(t, bodyStr, "/docs/api/swagger.json") // Check for updated swagger.json path
			},
		},
		{
			name: "serve root documentation",
			path: "/",
			setup: func(s *Server) {
				require.NoError(t, os.WriteFile(
					filepath.Join(s.docsPath, "README.md"),
					[]byte("# Root Documentation"),
					0644,
				))
			},
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				bodyStr := w.Body.String()
				assert.Contains(t, bodyStr, "<html>")
				assert.Contains(t, bodyStr, "# Root Documentation")
			},
		},
		{
			name: "handle non-existent file",
			path: "/non-existent.md",
			setup: func(_ *Server) {
				// No setup needed
			},
			expectedStatus: http.StatusNotFound,
			expectedType:   "text/plain; charset=utf-8",
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Documentation not found")
			},
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
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Diagram string `json:"diagram"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Contains(t, response.Diagram, "reader[FileReader]:::running")
				assert.Contains(t, response.Diagram, "writer[FileWriter]:::waiting")
				assert.Contains(t, response.Diagram, "reader -->|data| writer")
			},
		},
		{
			name: "serve docs through ServeHTTP",
			path: "/test-serve-http.md",
			setup: func(s *Server) {
				require.NoError(t, os.WriteFile(
					filepath.Join(s.docsPath, "test-serve-http.md"),
					[]byte("# ServeHTTP Test"),
					0644,
				))
			},
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "ServeHTTP Test")
			},
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

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
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

func TestServerServeHTTP(t *testing.T) {
	tmpDir := t.TempDir()
	srv := NewServer(Config{
		DocsPath: tmpDir,
	})
	srv.SetupRoutes()

	// Create test file
	testContent := "# Test Content"
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "test.md"),
		[]byte(testContent),
		0644,
	))

	// Test request
	req := httptest.NewRequest("GET", "/test.md", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), testContent)
}
