package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/elleshadow/noPromises/pkg/server/docs"
	"github.com/gorilla/mux"
)

// Config holds server configuration
type Config struct {
	Port     int
	DocsPath string
}

// Server represents the main server component
type Server struct {
	config     Config
	router     *mux.Router
	flows      *FlowManager
	processes  *ProcessRegistry
	docsServer *docs.Server
	Handler    http.Handler
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
	if config.DocsPath == "" {
		config.DocsPath = "docs"
	}

	s := &Server{
		config:    config,
		router:    mux.NewRouter(),
		flows:     newFlowManager(),
		processes: newProcessRegistry(),
	}

	s.setupRoutes()
	s.setupMiddleware()

	// Initialize docs server
	s.docsServer = docs.NewServer(docs.Config{
		DocsPath: config.DocsPath,
	})
	s.docsServer.SetupRoutes()

	// Mount docs routes under main router
	docsRouter := s.docsServer.Router()

	// API Documentation
	s.router.PathPrefix("/api-docs").Handler(docsRouter)
	s.router.PathPrefix("/api/swagger.json").Handler(docsRouter)

	// Network Diagrams
	s.router.PathPrefix("/diagrams/").Handler(docsRouter)

	// Documentation files
	s.router.PathPrefix("/docs/").Handler(docsRouter)

	// Root path redirects to docs
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})

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
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Flow management endpoints
	api.HandleFunc("/flows", s.handleCreateFlow).Methods(http.MethodPost)
	api.HandleFunc("/flows", s.handleListFlows).Methods(http.MethodGet)
	api.HandleFunc("/flows/{id}", s.handleGetFlow).Methods(http.MethodGet)
	api.HandleFunc("/flows/{id}", s.handleDeleteFlow).Methods(http.MethodDelete)
	api.HandleFunc("/flows/{id}/start", s.handleStartFlow).Methods(http.MethodPost)
	api.HandleFunc("/flows/{id}/stop", s.handleStopFlow).Methods(http.MethodPost)
	api.HandleFunc("/flows/{id}/status", s.handleGetFlowStatus).Methods(http.MethodGet)
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
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
	var flowConfig map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&flowConfig); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate flow configuration
	if err := s.validateFlowConfig(flowConfig); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	flowID := flowConfig["id"].(string)
	s.flows.mu.Lock()
	defer s.flows.mu.Unlock()

	// Check if flow already exists
	if _, exists := s.flows.flows[flowID]; exists {
		respondError(w, http.StatusConflict, fmt.Errorf("flow %s already exists", flowID))
		return
	}

	// Create new flow
	flow := &ManagedFlow{
		ID:     flowID,
		Config: flowConfig,
		State:  FlowStateCreated,
	}
	s.flows.flows[flowID] = flow

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

	// Create a copy of the flow for the response
	flowCopy := *flow
	s.flows.mu.Unlock()

	// Start flow in background
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.flows.mu.Lock()
		flow.State = FlowStateRunning
		s.flows.mu.Unlock()
	}()

	respondJSON(w, http.StatusOK, flowCopy)
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
	if config["id"] == nil {
		return fmt.Errorf("missing flow id")
	}

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

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.Handler,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	return srv.ListenAndServe()
}

// RegisterProcessType registers a new process type
func (s *Server) RegisterProcessType(name string, factory ProcessFactory) {
	s.processes.mu.Lock()
	defer s.processes.mu.Unlock()
	s.processes.processes[name] = factory
}
