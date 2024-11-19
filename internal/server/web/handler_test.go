package web

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestServer creates a server with test templates
func setupTestServer() *Server {
	// Create a basic template for testing
	tmpl := template.Must(template.New("index.html").Parse(`
		<!DOCTYPE html>
		<html>
			<head><title>{{.Title}}</title></head>
			<body>
				<h1>noPromises</h1>
				<p>Flow-Based Programming</p>
				<h2>Documentation</h2>
				<h2>Flow Management</h2>
			</body>
		</html>
	`))

	return NewServer(
		WithTemplates(tmpl),
		WithStatic(http.FileServer(http.Dir("testdata"))),
		WithFlowManager(&mockFlowManager{}),
	)
}

func TestHomeHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func() *Server
		expectedStatus int
		expectedBody   []string
	}{
		{
			name: "successful home page render",
			setupMock: func() *Server {
				return setupTestServer()
			},
			expectedStatus: http.StatusOK,
			expectedBody: []string{
				"noPromises",
				"Flow-Based Programming",
				"Documentation",
				"Flow Management",
			},
		},
		{
			name: "template not found",
			setupMock: func() *Server {
				return NewServer(WithTemplates(template.New("invalid")))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Internal Server Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupMock()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			server.HandleHome()(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := w.Body.String()
			for _, expected := range tt.expectedBody {
				if !strings.Contains(body, expected) {
					t.Errorf("expected body to contain %q", expected)
				}
			}
		})
	}
}

func TestStaticFileServer(t *testing.T) {
	// Create temporary test directory
	testDir := t.TempDir()

	// Create test files
	cssDir := filepath.Join(testDir, "css")
	jsDir := filepath.Join(testDir, "js")

	if err := os.MkdirAll(cssDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(cssDir, "style.css"), []byte("body {}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(jsDir, "main.js"), []byte("console.log('test');"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedTypes  []string
	}{
		{
			name:           "serve css file",
			path:           "/static/css/style.css",
			expectedStatus: http.StatusOK,
			expectedTypes:  []string{"text/css"},
		},
		{
			name:           "serve javascript file",
			path:           "/static/js/main.js",
			expectedStatus: http.StatusOK,
			expectedTypes:  []string{"application/javascript", "text/javascript"},
		},
		{
			name:           "file not found",
			path:           "/static/not-exists.txt",
			expectedStatus: http.StatusNotFound,
			expectedTypes:  []string{"text/plain"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a handler that strips the /static/ prefix
			fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir(testDir)))

			// Create server with the temporary test directory
			server := NewServer(
				WithTemplates(template.Must(template.New("index.html").Parse(`<html></html>`))),
				WithStatic(fileServer),
				WithFlowManager(&mockFlowManager{}),
			)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
				t.Logf("Request path: %s", tt.path)
				t.Logf("Test directory: %s", testDir)
			}

			contentType := w.Header().Get("Content-Type")
			validType := false
			for _, expectedType := range tt.expectedTypes {
				if strings.Contains(contentType, expectedType) {
					validType = true
					break
				}
			}
			if !validType {
				t.Errorf("content-type %q did not match any expected type: %v", contentType, tt.expectedTypes)
			}
		})
	}
}

// Mock flow manager for testing
type mockFlowManager struct{}

func (m *mockFlowManager) List() []ManagedFlow {
	return []ManagedFlow{
		{ID: "test-flow-1", Status: "running"},
		{ID: "test-flow-2", Status: "stopped"},
	}
}
