package docs

import (
	"fmt"
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
	// API documentation UI and swagger.json
	s.router.HandleFunc("/api-docs", s.handleSwaggerUI).Methods("GET")
	s.router.HandleFunc("/api/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, filepath.Join(s.docsPath, "api", "swagger.json"))
	}).Methods("GET")

	// Network visualization endpoints
	s.router.HandleFunc("/diagrams/network/{id}", s.handleNetworkDiagram).Methods("GET")
	s.router.HandleFunc("/diagrams/network/{id}/live", s.handleLiveDiagram).Methods("GET")

	// Serve static documentation files with HTML wrapper
	fileServer := http.FileServer(http.Dir(s.docsPath))
	s.router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Remove /docs prefix if present
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/docs")

		// If accessing root, serve README.md
		if r.URL.Path == "/" || r.URL.Path == "" {
			r.URL.Path = "/README.md"
		}

		// For Markdown files, wrap them in HTML
		if strings.HasSuffix(r.URL.Path, ".md") {
			content, err := os.ReadFile(filepath.Join(s.docsPath, r.URL.Path))
			if err != nil {
				http.Error(w, "Documentation not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			s.renderDocPage(w, string(content))
			return
		}

		// Serve other files normally
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

// Handler implementations...
