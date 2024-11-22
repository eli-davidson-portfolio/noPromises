package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/elleshadow/noPromises/internal/db"
	"github.com/gorilla/mux"
)

// Config holds server configuration
type Config struct {
	Port           int
	DocsPath       string
	DBPath         string
	MigrationsPath string // Path to migration files
}

// FlowManager manages flow instances
type FlowManager struct {
	flows map[string]*Flow
	mu    sync.RWMutex
}

// Flow represents a flow instance
type Flow struct {
	ID     string
	State  string
	Config map[string]interface{}
}

// Server represents the main server instance
type Server struct {
	config    Config
	router    *mux.Router
	flows     *FlowManager
	processes *ProcessRegistry
	webServer *http.Server
}

func newFlowManager() *FlowManager {
	return &FlowManager{
		flows: make(map[string]*Flow),
	}
}

// NewServer creates a new server instance
func NewServer(config Config) (*Server, error) {
	// Verify docs path exists
	if _, err := os.Stat(config.DocsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("docs path does not exist: %s", config.DocsPath)
	}

	// Verify required files exist
	requiredFiles := []string{
		"README.md",
		"api/swagger.json",
	}
	for _, file := range requiredFiles {
		path := filepath.Join(config.DocsPath, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("required file missing: %s", path)
		}
	}

	// Initialize database
	database, err := db.New(config.DBPath)
	if err != nil {
		return nil, fmt.Errorf("initializing database: %w", err)
	}

	// Run migrations
	migrationManager := db.NewMigrationManager(database)
	migrationsPath := config.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "internal/db/migrations"
	}
	if err := migrationManager.ApplyMigrations(context.Background(), migrationsPath); err != nil {
		return nil, fmt.Errorf("applying migrations: %w", err)
	}

	// Get current version for logging
	version, err := migrationManager.GetCurrentVersion(context.Background())
	if err != nil {
		log.Printf("[WARN] Failed to get current schema version: %v", err)
	} else {
		log.Printf("[INFO] Database initialized with schema version %d", version)
	}

	s := &Server{
		config:    config,
		router:    mux.NewRouter(),
		flows:     newFlowManager(),
		processes: newProcessRegistry(),
	}

	s.setupRoutes()
	return s, nil
}

func (s *Server) setupRoutes() {
	// Serve static files from docs directory
	s.router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir(s.config.DocsPath))))

	// Serve API documentation
	s.router.PathPrefix("/api/").Handler(http.StripPrefix("/api/", http.FileServer(http.Dir(filepath.Join(s.config.DocsPath, "api")))))

	// Serve home page
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(s.config.DocsPath, "README.md"))
	})
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	s.webServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.router,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.webServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return s.webServer.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

// CreateFlow creates a new flow
func (s *Server) CreateFlow(id string, config map[string]interface{}) error {
	s.flows.mu.Lock()
	defer s.flows.mu.Unlock()

	s.flows.flows[id] = &Flow{
		ID:     id,
		State:  "created",
		Config: config,
	}
	return nil
}

// GetFlow retrieves a flow by ID
func (s *Server) GetFlow(id string) (*Flow, bool) {
	s.flows.mu.RLock()
	defer s.flows.mu.RUnlock()

	flow, exists := s.flows.flows[id]
	return flow, exists
}

// StartFlow starts a flow
func (s *Server) StartFlow(_ context.Context, id string) error {
	s.flows.mu.Lock()
	defer s.flows.mu.Unlock()

	flow, exists := s.flows.flows[id]
	if !exists {
		return fmt.Errorf("flow not found: %s", id)
	}

	flow.State = "running"
	return nil
}

// StopFlow stops a flow
func (s *Server) StopFlow(_ context.Context, id string) error {
	s.flows.mu.Lock()
	defer s.flows.mu.Unlock()

	flow, exists := s.flows.flows[id]
	if !exists {
		return fmt.Errorf("flow not found: %s", id)
	}

	flow.State = "stopped"
	return nil
}

// RegisterProcessType registers a new process type
func (s *Server) RegisterProcessType(name string, factory ProcessFactory) {
	s.processes.Register(name, factory)
}
