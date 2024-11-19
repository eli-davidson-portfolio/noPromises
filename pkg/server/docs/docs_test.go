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
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "test.md"),
		[]byte("# Test Documentation"),
		0644,
	))

	// Create docs server
	srv := NewServer(Config{
		DocsPath: tmpDir,
	})
	require.NotNil(t, srv)

	// Setup routes
	srv.SetupRoutes()

	// Setup test cases
	tests := []struct {
		name           string
		path           string
		setup          func(*Server)
		expectedStatus int
		expectedType   string
	}{
		{
			name: "serve static documentation",
			path: "/docs/test.md",
			setup: func(_ *Server) {
				// No setup needed for static files
			},
			expectedStatus: http.StatusOK,
			expectedType:   "text/markdown; charset=utf-8",
		},
		{
			name: "generate network diagram",
			path: "/diagrams/network/test-flow",
			setup: func(s *Server) {
				// Setup test network data
				s.mermaidGen.SetNetwork("test-flow", map[string]interface{}{
					"nodes": map[string]interface{}{
						"reader": map[string]interface{}{
							"type":   "FileReader",
							"status": "running",
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
			// Setup test case
			if tt.setup != nil {
				tt.setup(srv)
			}

			// Create test request
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			// Handle request
			srv.Router().ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedType, w.Header().Get("Content-Type"))

			// For diagram tests, verify response body contains expected elements
			if tt.path == "/diagrams/network/test-flow" {
				assert.Contains(t, w.Body.String(), "reader[FileReader]:::running")
			}
		})
	}
}

func TestNetworkDiagramGeneration(t *testing.T) {
	srv := NewServer(Config{
		DocsPath: "testdata/docs",
	})
	srv.SetupRoutes()

	// Setup test network
	testNetwork := map[string]interface{}{
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
	}

	srv.mermaidGen.SetNetwork("test-flow", testNetwork)

	// Test diagram generation
	req := httptest.NewRequest("GET", "/diagrams/network/test-flow", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response JSON
	var response struct {
		Diagram string `json:"diagram"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Split diagram into lines for easier testing
	lines := strings.Split(response.Diagram, "\n")
	var nodeLines, edgeLines, styleLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.Contains(line, "["):
			nodeLines = append(nodeLines, line)
		case strings.Contains(line, "-->"):
			edgeLines = append(edgeLines, line)
		case strings.Contains(line, "classDef"):
			styleLines = append(styleLines, line)
		}
	}

	// Test node definitions
	assert.Contains(t, nodeLines, "reader[FileReader]:::running")
	assert.Contains(t, nodeLines, "writer[FileWriter]:::waiting")

	// Test edge definition
	assert.Contains(t, edgeLines, "reader -->|data| writer")

	// Test style definitions
	assert.Contains(t, styleLines, "classDef running fill:#d4edda,stroke:#28a745;")
	assert.Contains(t, styleLines, "classDef waiting fill:#fff3cd,stroke:#ffc107;")
	assert.Contains(t, styleLines, "classDef error fill:#f8d7da,stroke:#dc3545;")
}

func TestLiveUpdates(t *testing.T) {
	srv := NewServer(Config{
		DocsPath: "testdata/docs",
	})
	srv.SetupRoutes()

	// Test WebSocket connection
	req := httptest.NewRequest("GET", "/diagrams/network/test-flow/live", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusSwitchingProtocols, w.Code)
}
