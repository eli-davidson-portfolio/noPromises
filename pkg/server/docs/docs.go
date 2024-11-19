package docs

import (
	"net/http"
	"path/filepath"

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
	// Serve static documentation files with proper content types
	fileServer := http.FileServer(http.Dir(s.docsPath))
	s.router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		fileServer.ServeHTTP(w, r)
	})))

	// Network visualization endpoints
	s.router.HandleFunc("/diagrams/network/{id}", s.handleNetworkDiagram)
	s.router.HandleFunc("/diagrams/network/{id}/live", s.handleLiveDiagram)

	// API documentation
	s.router.HandleFunc("/api-docs", s.handleSwaggerUI)
}

// Handler implementations...
