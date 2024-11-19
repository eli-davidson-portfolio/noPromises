package server

import (
	"net/http"

	"github.com/elleshadow/noPromises/internal/server/web"
)

// Server represents the main server instance
type Server struct {
	router *http.ServeMux
	web    *web.Server
	flows  web.FlowManager
}

// NewServer creates a new server instance
func NewServer(flows web.FlowManager) *Server {
	router := http.NewServeMux()
	webServer := web.NewServer(
		web.WithFlowManager(flows),
	)

	s := &Server{
		router: router,
		web:    webServer,
		flows:  flows,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all server routes
func (s *Server) setupRoutes() {
	// Mount web interface handlers
	s.router.HandleFunc("/", s.web.ServeHTTP)         // Handle root explicitly
	s.router.Handle("/static/", s.web)                // Static files
	s.router.Handle("/api/v1/flows", s.web)           // Flow API
	s.router.Handle("/api/v1/flows/", s.web)          // Flow API with IDs
	s.router.Handle("/docs/", http.NotFoundHandler()) // Temporarily disable docs redirect
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
