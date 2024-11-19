package docs

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

type Config struct {
	DocsPath string
}

type Server struct {
	router     *mux.Router
	docsPath   string
	mermaidGen *MermaidGenerator
}

func NewServer(config Config) *Server {
	return &Server{
		router:     mux.NewRouter(),
		docsPath:   config.DocsPath,
		mermaidGen: NewMermaidGenerator(),
	}
}

func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) SetupRoutes() {
	s.logDebug("Setting up docs server routes with docsPath: %s", s.docsPath)

	// API documentation UI and swagger.json
	s.router.HandleFunc("/api-docs", s.HandleSwaggerUI).Methods("GET")
	s.router.HandleFunc("/api/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		s.logDebug("Serving swagger.json from: %s", filepath.Join(s.docsPath, "api", "swagger.json"))
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, filepath.Join(s.docsPath, "api", "swagger.json"))
	}).Methods("GET")

	// Network visualization endpoints
	s.router.HandleFunc("/diagrams/network/{id}", s.handleNetworkDiagram).Methods("GET")
	s.router.HandleFunc("/diagrams/network/{id}/live", s.handleLiveDiagram).Methods("GET")

	// Serve static documentation files with HTML wrapper
	fileServer := http.FileServer(http.Dir(s.docsPath))
	s.router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logDebug("Incoming request path: %s", r.URL.Path)
		s.logDebug("Looking in directory: %s", s.docsPath)

		// Remove /docs prefix if present
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/docs")
		s.logDebug("Path after trim: %s", r.URL.Path)

		// If accessing root, serve README.md
		if r.URL.Path == "/" || r.URL.Path == "" {
			r.URL.Path = "/README.md"
			s.logDebug("Serving root, updated path to: %s", r.URL.Path)
		}

		fullPath := filepath.Join(s.docsPath, r.URL.Path)
		s.logDebug("Full file path: %s", fullPath)

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			s.logDebug("File not found: %s", fullPath)
			http.Error(w, "Documentation not found", http.StatusNotFound)
			return
		}

		// For Markdown files, wrap them in HTML
		if strings.HasSuffix(r.URL.Path, ".md") {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				s.logDebug("Error reading file: %v", err)
				http.Error(w, "Documentation not found", http.StatusNotFound)
				return
			}

			s.logDebug("Serving markdown file with HTML wrapper")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			s.renderDocPage(w, string(content))
			return
		}

		// Serve other files normally
		s.logDebug("Serving static file")
		fileServer.ServeHTTP(w, r)
	}))
}

// renderDocPage wraps markdown content in a styled HTML page
func (s *Server) renderDocPage(w http.ResponseWriter, content string) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>noPromises Documentation</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/github-markdown-css@5/github-markdown.min.css">
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <style>
        body {
            box-sizing: border-box;
            min-width: 200px;
            max-width: 980px;
            margin: 0 auto;
            padding: 45px;
            background: #f6f8fa;
        }
        .markdown-body {
            background: white;
            padding: 45px;
            border-radius: 6px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.12);
        }
        @media (max-width: 767px) {
            body { padding: 15px; }
            .markdown-body { padding: 15px; }
        }
        nav {
            margin-bottom: 20px;
            padding: 10px;
            background: white;
            border-radius: 6px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.12);
        }
        nav a {
            color: #0366d6;
            text-decoration: none;
            margin-right: 15px;
        }
        nav a:hover { text-decoration: underline; }
        pre { background: #f6f8fa; padding: 16px; border-radius: 6px; }
        code { font-family: SFMono-Regular,Consolas,Liberation Mono,Menlo,monospace; }
    </style>
</head>
<body>
    <nav>
        <a href="/docs">Home</a>
        <a href="/docs/guides/getting-started.md">Getting Started</a>
        <a href="/docs/api/endpoints.md">API</a>
        <a href="/docs/architecture">Architecture</a>
    </nav>
    <div class="markdown-body">
        <div id="content"></div>
    </div>
    <script>
        // Render markdown content
        document.getElementById('content').innerHTML = marked.parse(` + "`" + content + "`" + `);

        // Add syntax highlighting to code blocks
        document.querySelectorAll('pre code').forEach(block => {
            block.className = 'language-' + (block.className || 'plaintext');
        });
    </script>
</body>
</html>`

	fmt.Fprint(w, html)
}

// Add these debug logging functions
func (s *Server) logDebug(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

// Handler implementations...

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.logDebug("Serving docs request: %s", r.URL.Path)
	s.router.ServeHTTP(w, r)
}

// Add this exported method
func (s *Server) HandleSwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte(`<!DOCTYPE html>
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
</html>`))
	if err != nil {
		s.logDebug("Error writing Swagger UI response: %v", err)
	}
}
