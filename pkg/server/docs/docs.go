package docs

import (
	"net/http"
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

	// Serve static documentation files with proper content types
	fileServer := http.FileServer(http.Dir(s.docsPath))
	s.router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Remove /docs prefix if present (TrimPrefix is safe to use unconditionally)
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/docs")

		// Set content type based on file extension
		ext := filepath.Ext(r.URL.Path)
		switch ext {
		case ".md":
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		default:
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}

		// If accessing root, serve README.md
		if r.URL.Path == "/" || r.URL.Path == "" {
			r.URL.Path = "/README.md"
		}

		fileServer.ServeHTTP(w, r)
	}))
}

// Handler implementations...
