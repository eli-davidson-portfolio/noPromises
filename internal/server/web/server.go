package web

import (
	"html/template"
	"net/http"
	"time"
)

// ManagedFlow represents a flow with its runtime state
type ManagedFlow struct {
	ID        string                 `json:"id"`
	Config    map[string]interface{} `json:"config"`
	State     string                 `json:"state"`
	StartTime *time.Time             `json:"started_at,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// FlowManager defines the interface for managing flows
type FlowManager interface {
	List() []ManagedFlow
	Get(id string) (*ManagedFlow, bool)
}

// Server handles web interface requests
type Server struct {
	flowManager FlowManager
	templates   *template.Template
}

// ServerOption configures the web server
type ServerOption func(*Server)

// WithFlowManager sets the flow manager
func WithFlowManager(fm FlowManager) ServerOption {
	return func(s *Server) {
		s.flowManager = fm
	}
}

// WithTemplates sets the HTML templates
func WithTemplates(t *template.Template) ServerOption {
	return func(s *Server) {
		s.templates = t
	}
}

// NewServer creates a new web server
func NewServer(opts ...ServerOption) *Server {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// HandleHome renders the home page
func (s *Server) HandleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		data := struct {
			Title string
		}{
			Title: "NoPromises Flow Manager",
		}
		if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// HandleFlows renders the flows page
func (s *Server) HandleFlows() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		flows := s.flowManager.List()
		data := struct {
			Title   string
			Flows   []ManagedFlow
			Content string
		}{
			Title:   "Active Flows",
			Flows:   flows,
			Content: "", // Empty content for now
		}
		err := s.templates.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Add this method to implement http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		s.HandleHome()(w, r)
	case "/flows":
		s.HandleFlows()(w, r)
	default:
		http.NotFound(w, r)
	}
}
