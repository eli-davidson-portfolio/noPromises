package docs

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday/v2"
)

// Config holds documentation server configuration
type Config struct {
	DocsPath string
}

// Server handles documentation serving
type Server struct {
	router     *mux.Router
	docsPath   string
	mermaidGen *MermaidGenerator
}

// NewServer creates a new documentation server
func NewServer(config Config) *Server {
	return &Server{
		router:     mux.NewRouter(),
		docsPath:   config.DocsPath,
		mermaidGen: NewMermaidGenerator(),
	}
}

// Router returns the server's router
func (s *Server) Router() *mux.Router {
	return s.router
}

// SetupRoutes configures the server routes
func (s *Server) SetupRoutes() {
	if s.router == nil {
		s.router = mux.NewRouter()
	}

	// Serve static files first - use the docs path for static files
	staticPath := filepath.Join(s.docsPath, "static")
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir(staticPath))))

	// API documentation
	s.router.HandleFunc("/api-docs", s.HandleSwaggerUI)
	s.router.HandleFunc("/docs/api/swagger.json", s.HandleSwaggerJSON)

	// Add diagram routes
	s.router.HandleFunc("/diagrams/network/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		s.HandleFlowDiagram(w, r, vars["id"])
	})

	s.router.HandleFunc("/diagrams/network/{id}/live", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusSwitchingProtocols)
	})

	// Handle markdown files - must be last to avoid conflicting with other routes
	s.router.PathPrefix("/docs/").HandlerFunc(s.HandleMarkdown)
}

// HandleMarkdown renders markdown files as HTML
func (s *Server) HandleMarkdown(w http.ResponseWriter, r *http.Request) {
	s.logDebug("Received markdown request for: %s", r.URL.Path)

	filePath := filepath.Join(s.docsPath, strings.TrimPrefix(r.URL.Path, "/docs/"))
	s.logDebug("Looking for file at: %s", filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		s.logDebug("File read error: %v", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}
	s.logDebug("Successfully read file of size: %d bytes", len(content))

	// Convert markdown to HTML
	html := blackfriday.Run(content)
	s.logDebug("Converted to HTML of size: %d bytes", len(html))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	s.renderDocPage(w, string(html))
	s.logDebug("Response sent")
}

// HandleSwaggerJSON serves the OpenAPI specification
func (s *Server) HandleSwaggerJSON(w http.ResponseWriter, _ *http.Request) {
	filePath := filepath.Join(s.docsPath, "api", "swagger.json")
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Swagger specification not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(content); err != nil {
		s.logDebug("Error writing swagger response: %v", err)
	}
}

// HandleSwaggerUI serves the Swagger UI
func (s *Server) HandleSwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
  <head>
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
    <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/favicon-16x16.png" sizes="16x16" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
      window.onload = function() {
        const ui = SwaggerUIBundle({
          url: "/docs/api/swagger.json",
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
          ],
          plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
          ],
          layout: "StandaloneLayout"
        });
        window.ui = ui;
      }
    </script>
  </body>
</html>`)
}

// renderDocPage wraps HTML content in a styled page
func (s *Server) renderDocPage(w http.ResponseWriter, content string) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>noPromises Documentation</title>
    <link rel="stylesheet" href="/static/css/github-markdown.css">
    <style>
        .markdown-body {
            box-sizing: border-box;
            min-width: 200px;
            max-width: 980px;
            margin: 0 auto;
            padding: 45px;
        }
    </style>
</head>
<body class="markdown-body">
    %s
</body>
</html>`

	fmt.Fprintf(w, tmpl, content)
}

func (s *Server) logDebug(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
