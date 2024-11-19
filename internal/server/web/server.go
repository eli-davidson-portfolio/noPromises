package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// ManagedFlow represents a flow in the system
type ManagedFlow struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// FlowManager interface for managing flows
type FlowManager interface {
	List() []ManagedFlow
}

// Server handles web interface requests
type Server struct {
	templates *template.Template
	flows     FlowManager
	static    http.Handler
}

// NewServer creates a new web interface server
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		flows: &defaultFlowManager{},
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Set defaults if not provided through options
	if s.templates == nil {
		tmpl := template.Must(template.ParseGlob("web/templates/*.html"))
		s.templates = tmpl
	}
	if s.static == nil {
		s.static = http.FileServer(http.Dir("web/static"))
	}

	return s
}

// ServerOption allows customizing the server
type ServerOption func(*Server)

// WithTemplates sets custom templates for the server
func WithTemplates(t *template.Template) ServerOption {
	return func(s *Server) {
		s.templates = t
	}
}

// WithStatic sets custom static file handler
func WithStatic(h http.Handler) ServerOption {
	return func(s *Server) {
		s.static = h
	}
}

// WithFlowManager sets custom flow manager
func WithFlowManager(fm FlowManager) ServerOption {
	return func(s *Server) {
		s.flows = fm
	}
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/static/") {
		s.static.ServeHTTP(w, r)
		return
	}

	switch r.URL.Path {
	case "/":
		s.HandleHome()(w, r)
	case "/api/v1/flows":
		s.HandleFlows()(w, r)
	default:
		if strings.HasPrefix(r.URL.Path, "/api/v1/flows/") && strings.HasSuffix(r.URL.Path, "/viz") {
			s.HandleFlowVisualization()(w, r)
			return
		}
		http.NotFound(w, r)
	}
}

// HandleHome returns the handler for the home page
func (s *Server) HandleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if s.templates == nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := struct {
			Title string
			Flows []ManagedFlow
		}{
			Title: "noPromises Dashboard",
			Flows: s.flows.List(),
		}

		err := s.templates.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// HandleFlows returns the handler for flow management
func (s *Server) HandleFlows() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			flows := s.flows.List()
			w.Header().Set("Content-Type", "text/html")
			for _, flow := range flows {
				fmt.Fprintf(w, `<div class="flow-item">%s</div>`, flow.ID)
			}

		case http.MethodPost:
			var newFlow struct {
				ID     string         `json:"id"`
				Config map[string]any `json:"config"`
			}
			if err := json.NewDecoder(r.Body).Decode(&newFlow); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<div class="flow-item">%s</div>`, newFlow.ID)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// HandleFlowVisualization returns the handler for flow visualization
func (s *Server) HandleFlowVisualization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flowID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/flows/"), "/viz")

		// Check if flow exists
		found := false
		for _, flow := range s.flows.List() {
			if flow.ID == flowID {
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Flow not found", http.StatusNotFound)
			return
		}

		// Return mock visualization data
		response := struct {
			Nodes []map[string]string `json:"nodes"`
			Edges []map[string]string `json:"edges"`
		}{
			Nodes: []map[string]string{
				{"id": "1", "label": "Start"},
				{"id": "2", "label": "End"},
			},
			Edges: []map[string]string{
				{"from": "1", "to": "2"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// defaultFlowManager is a basic implementation of FlowManager
type defaultFlowManager struct{}

func (m *defaultFlowManager) List() []ManagedFlow {
	return []ManagedFlow{
		{ID: "test-flow-1", Status: "running"},
		{ID: "test-flow-2", Status: "stopped"},
	}
}
