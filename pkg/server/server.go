package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/elleshadow/noPromises/internal/server/web"
	"github.com/gorilla/mux"
)

// Config holds server configuration
type Config struct {
	Port     int
	DocsPath string
}

// Server represents the main server component
type Server struct {
	config    Config
	router    *mux.Router
	flows     *FlowManager
	processes *ProcessRegistry
	webServer *web.Server
	Handler   http.Handler
}

// FlowManager handles flow lifecycle and state management
type FlowManager struct {
	flows map[string]*ManagedFlow
	mu    sync.RWMutex
}

// ManagedFlow represents a flow with its runtime state
type ManagedFlow struct {
	ID        string                 `json:"id"`
	Config    map[string]interface{} `json:"config"`
	State     FlowState              `json:"state"`
	StartTime *time.Time             `json:"started_at,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// FlowState represents the possible states of a flow
type FlowState string

const (
	FlowStateCreated  FlowState = "created"
	FlowStateStarting FlowState = "starting"
	FlowStateRunning  FlowState = "running"
	FlowStateStopping FlowState = "stopping"
	FlowStateStopped  FlowState = "stopped"
	FlowStateError    FlowState = "error"
)

// ProcessRegistry manages available process types
type ProcessRegistry struct {
	processes map[string]ProcessFactory
	mu        sync.RWMutex
}

// ProcessFactory creates new process instances
type ProcessFactory interface {
	Create(config map[string]interface{}) (Process, error)
}

// Process represents a flow process
type Process interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// NewServer creates a new server instance
func NewServer(config Config) (*Server, error) {
	flowManager := newFlowManager()

	s := &Server{
		config:    config,
		router:    mux.NewRouter(),
		flows:     flowManager,
		processes: newProcessRegistry(),
		webServer: web.NewServer(
			web.WithFlowManager(flowManager),
		),
	}

	s.setupRoutes()
	s.setupMiddleware()

	s.Handler = s.router
	return s, nil
}

func newFlowManager() *FlowManager {
	return &FlowManager{
		flows: make(map[string]*ManagedFlow),
	}
}

func newProcessRegistry() *ProcessRegistry {
	return &ProcessRegistry{
		processes: make(map[string]ProcessFactory),
	}
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/flows", s.handleCreateFlow).Methods(http.MethodPost)
	api.HandleFunc("/flows", s.handleListFlows).Methods(http.MethodGet)
	api.HandleFunc("/flows/{id}", s.handleGetFlow).Methods(http.MethodGet)
	api.HandleFunc("/flows/{id}", s.handleDeleteFlow).Methods(http.MethodDelete)
	api.HandleFunc("/flows/{id}/start", s.handleStartFlow).Methods(http.MethodPost)
	api.HandleFunc("/flows/{id}/stop", s.handleStopFlow).Methods(http.MethodPost)
	api.HandleFunc("/flows/{id}/status", s.handleGetFlowStatus).Methods(http.MethodGet)

	// Static files - handle before the catch-all route
	staticDir := filepath.Join("web", "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = filepath.Join(s.config.DocsPath, "static")
	}
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
	s.router.PathPrefix("/static/").Handler(staticHandler)

	// Web interface (must be last as it's the catch-all)
	s.router.PathPrefix("/").Handler(s.webServer)
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only set JSON content type for API routes
			if strings.HasPrefix(r.URL.Path, "/api/v1/") {
				w.Header().Set("Content-Type", "application/json")
			}
			next.ServeHTTP(w, r)
		})
	})
}

// Response helpers
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": data,
		}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

func respondError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": err.Error(),
		},
	}); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}

// Flow management handlers
func (s *Server) handleCreateFlow(w http.ResponseWriter, r *http.Request) {
	var flowConfig struct {
		ID     string                 `json:"id"`
		Config map[string]interface{} `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&flowConfig); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate flow configuration
	if err := s.validateFlowConfig(flowConfig.Config); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	s.flows.mu.Lock()
	defer s.flows.mu.Unlock()

	// Check if flow already exists
	if _, exists := s.flows.flows[flowConfig.ID]; exists {
		respondError(w, http.StatusConflict, fmt.Errorf("flow %s already exists", flowConfig.ID))
		return
	}

	// Create new flow
	flow := &ManagedFlow{
		ID:     flowConfig.ID,
		Config: flowConfig.Config,
		State:  FlowStateCreated,
	}
	s.flows.flows[flowConfig.ID] = flow

	respondJSON(w, http.StatusCreated, flow)
}

func (s *Server) handleGetFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowID := vars["id"]

	s.flows.mu.RLock()
	flow, exists := s.flows.flows[flowID]
	s.flows.mu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, fmt.Errorf("flow %s not found", flowID))
		return
	}

	respondJSON(w, http.StatusOK, flow)
}

func (s *Server) handleListFlows(w http.ResponseWriter, _ *http.Request) {
	s.flows.mu.RLock()
	flows := make([]*ManagedFlow, 0, len(s.flows.flows))
	for _, flow := range s.flows.flows {
		flows = append(flows, flow)
	}
	s.flows.mu.RUnlock()

	respondJSON(w, http.StatusOK, flows)
}

func (s *Server) handleStartFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowID := vars["id"]

	s.flows.mu.Lock()
	flow, exists := s.flows.flows[flowID]
	if !exists {
		s.flows.mu.Unlock()
		respondError(w, http.StatusNotFound, fmt.Errorf("flow %s not found", flowID))
		return
	}

	if flow.State == FlowStateRunning {
		s.flows.mu.Unlock()
		respondError(w, http.StatusConflict, fmt.Errorf("flow %s is already running", flowID))
		return
	}

	flow.State = FlowStateStarting
	now := time.Now()
	flow.StartTime = &now
	s.flows.mu.Unlock()

	// Start flow in background
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.flows.mu.Lock()
		flow.State = FlowStateRunning
		s.flows.mu.Unlock()
	}()

	respondJSON(w, http.StatusOK, flow)
}

func (s *Server) handleStopFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowID := vars["id"]

	s.flows.mu.Lock()
	flow, exists := s.flows.flows[flowID]
	if !exists {
		s.flows.mu.Unlock()
		respondError(w, http.StatusNotFound, fmt.Errorf("flow %s not found", flowID))
		return
	}

	if flow.State != FlowStateRunning {
		s.flows.mu.Unlock()

		respondError(w, http.StatusConflict, fmt.Errorf("flow %s is not running", flowID))
		return
	}

	flow.State = FlowStateStopping
	s.flows.mu.Unlock()

	// Stop flow in background
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.flows.mu.Lock()
		flow.State = FlowStateStopped
		s.flows.mu.Unlock()
	}()

	respondJSON(w, http.StatusOK, flow)
}

func (s *Server) handleDeleteFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowID := vars["id"]

	s.flows.mu.Lock()
	flow, exists := s.flows.flows[flowID]
	if !exists {
		s.flows.mu.Unlock()
		respondError(w, http.StatusNotFound, fmt.Errorf("flow %s not found", flowID))
		return
	}

	if flow.State == FlowStateRunning {
		s.flows.mu.Unlock()
		respondError(w, http.StatusConflict, fmt.Errorf("cannot delete running flow %s", flowID))
		return
	}

	delete(s.flows.flows, flowID)
	s.flows.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetFlowStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowID := vars["id"]

	s.flows.mu.RLock()
	flow, exists := s.flows.flows[flowID]
	s.flows.mu.RUnlock()

	if !exists {
		respondError(w, http.StatusNotFound, fmt.Errorf("flow %s not found", flowID))
		return
	}

	respondJSON(w, http.StatusOK, flow)
}

// Validation helpers
func (s *Server) validateFlowConfig(config map[string]interface{}) error {
	nodes, ok := config["nodes"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid nodes configuration")
	}

	for _, node := range nodes {
		nodeConfig, ok := node.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid node configuration")
		}

		nodeType, ok := nodeConfig["type"].(string)
		if !ok {
			return fmt.Errorf("missing node type")
		}

		if !s.isValidProcessType(nodeType) {
			return fmt.Errorf("invalid process type: %s", nodeType)
		}
	}

	return nil
}

func (s *Server) isValidProcessType(processType string) bool {
	s.processes.mu.RLock()
	defer s.processes.mu.RUnlock()
	_, exists := s.processes.processes[processType]
	return exists
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Handler.ServeHTTP(w, r)
}

// RegisterProcessType registers a new process type
func (s *Server) RegisterProcessType(name string, factory ProcessFactory) {
	s.processes.mu.Lock()
	defer s.processes.mu.Unlock()
	s.processes.processes[name] = factory
}

// Make FlowManager implement web.FlowManager interface
func (fm *FlowManager) List() []web.ManagedFlow {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	flows := make([]web.ManagedFlow, 0, len(fm.flows))
	for _, flow := range fm.flows {
		flows = append(flows, web.ManagedFlow{
			ID:     flow.ID,
			Status: string(flow.State),
		})
	}
	return flows
}

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.Handler,
	}

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	log.Printf("Server starting on http://localhost:%d", s.config.Port)
	return srv.ListenAndServe()
}
