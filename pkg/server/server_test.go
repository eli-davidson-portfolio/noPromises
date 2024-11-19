package server

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elleshadow/noPromises/internal/server/web"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a server with test configuration
func setupTestServer(t *testing.T) (*Server, string) {
	// Create test directory structure
	testDir := t.TempDir()
	webDir := filepath.Join(testDir, "web")
	staticDir := filepath.Join(webDir, "static")
	templatesDir := filepath.Join(webDir, "templates")

	require.NoError(t, os.MkdirAll(filepath.Join(staticDir, "css"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(staticDir, "js"), 0755))
	require.NoError(t, os.MkdirAll(templatesDir, 0755))

	// Create test files
	require.NoError(t, os.WriteFile(
		filepath.Join(staticDir, "css", "style.css"),
		[]byte("body {}"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(staticDir, "js", "main.js"),
		[]byte("console.log('test');"),
		0644,
	))

	// Create test template
	tmpl := template.Must(template.New("index.html").Parse(`
		<!DOCTYPE html>
		<html>
			<head><title>{{.Title}}</title></head>
			<body>
				<h1>noPromises</h1>
				<p>Flow-Based Programming</p>
			</body>
		</html>
	`))

	// Create test web server
	webServer := web.NewServer(
		web.WithTemplates(tmpl),
		web.WithStatic(http.FileServer(http.Dir(staticDir))),
	)

	s := &Server{
		config: Config{
			Port:     8080,
			DocsPath: webDir,
		},
		router:    mux.NewRouter(),
		flows:     newFlowManager(),
		processes: newProcessRegistry(),
		webServer: webServer,
	}

	s.setupRoutes()
	s.setupMiddleware()
	s.Handler = s.router

	return s, testDir
}

// setupTestServerWithoutWeb creates a server without web interface for process registry tests
func setupTestServerWithoutWeb(_ *testing.T) *Server {
	s := &Server{
		config:    Config{Port: 8080},
		router:    mux.NewRouter(),
		flows:     newFlowManager(),
		processes: newProcessRegistry(),
	}

	s.Handler = s.router
	return s
}

func TestNewServer(t *testing.T) {
	srv, _ := setupTestServer(t)
	require.NotNil(t, srv)
	require.NotNil(t, srv.router)
	require.NotNil(t, srv.flows)
	require.NotNil(t, srv.processes)
	require.NotNil(t, srv.webServer)
}

func TestServerRoutes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "home page",
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedType:   "text/html",
		},
		{
			name:           "static file",
			method:         http.MethodGet,
			path:           "/static/css/style.css",
			expectedStatus: http.StatusOK,
			expectedType:   "text/css",
		},
		{
			name:           "list flows API",
			method:         http.MethodGet,
			path:           "/api/v1/flows",
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
		{
			name:           "create flow API",
			method:         http.MethodPost,
			path:           "/api/v1/flows",
			body:           `{"id":"test-flow","config":{"nodes":{"test":{"type":"test"}}}}`,
			expectedStatus: http.StatusCreated,
			expectedType:   "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := setupTestServer(t)

			// Register test process type
			srv.RegisterProcessType("test", &mockProcessFactory{})

			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
				t.Logf("Request path: %s", tt.path)
				t.Logf("Response body: %s", w.Body.String())
			}

			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, tt.expectedType) {
				t.Errorf("expected content-type to contain %q, got %q", tt.expectedType, contentType)
			}
		})
	}
}

func TestFlowManagement(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Register test process type
	srv.RegisterProcessType("test", &mockProcessFactory{})

	// Test flow creation
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/flows", strings.NewReader(`
		{"id": "test-flow", "config": {"nodes": {"test": {"type": "test"}}}}
	`))
	createReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, createReq)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Test flow listing
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/flows", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, listReq)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test-flow")
}

func TestProcessRegistry(t *testing.T) {
	srv := setupTestServerWithoutWeb(t)

	// Register a test process
	srv.RegisterProcessType("test", &mockProcessFactory{})

	// Verify process type is registered
	srv.processes.mu.RLock()
	_, exists := srv.processes.processes["test"]
	srv.processes.mu.RUnlock()
	assert.True(t, exists)
}

// Mock implementations for testing
type mockProcessFactory struct{}

func (f *mockProcessFactory) Create(_ map[string]interface{}) (Process, error) {
	return &mockProcess{}, nil
}

type mockProcess struct{}

func (p *mockProcess) Start(_ context.Context) error { return nil }
func (p *mockProcess) Stop(_ context.Context) error  { return nil }
