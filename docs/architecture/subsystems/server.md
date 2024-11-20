# Server Architecture

## Core Components

### Server Structure
```go
type Server struct {
    config    Config           // Server configuration
    router    *mux.Router      // HTTP router
    flows     *FlowManager     // Flow management
    processes *ProcessRegistry // Process type registry
    webServer *http.Server     // HTTP server
}
```

### Configuration
```go
type Config struct {
    Port     int    // HTTP server port
    DocsPath string // Path to documentation files
    DBPath   string // Database path (optional)
}
```

## Flow Management

### Flow Manager
```go
type FlowManager struct {
    flows map[string]*Flow
    mu    sync.RWMutex
}

type Flow struct {
    ID     string
    State  string                 // created, running, stopped
    Config map[string]interface{} // Flow configuration
}
```

### Process Registry
```go
type ProcessRegistry struct {
    processes map[string]ProcessFactory
    mu        sync.RWMutex
}

func (r *ProcessRegistry) Register(typeName string, factory ProcessFactory) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.processes[typeName] = factory
}
```

## Route Handling

### Documentation Routes
```go
// Serve static documentation
s.router.PathPrefix("/docs/").Handler(
    http.StripPrefix("/docs/", 
    http.FileServer(http.Dir(s.config.DocsPath))))

// Serve API documentation
s.router.PathPrefix("/api/").Handler(
    http.StripPrefix("/api/", 
    http.FileServer(http.Dir(filepath.Join(s.config.DocsPath, "api")))))

// Serve home page
s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, filepath.Join(s.config.DocsPath, "README.md"))
})
```

## Server Lifecycle

### Initialization
1. Verify documentation path exists
2. Check for required files (README.md, swagger.json)
3. Create router and managers
4. Set up routes

### Server Start
```go
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
```

## Flow Operations

### Flow Creation
```go
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
```

### Flow Control
- Start Flow: Changes flow state to "running"
- Stop Flow: Changes flow state to "stopped"
- Get Flow: Retrieves flow by ID
- Flow State Management: Thread-safe state transitions

## Process Management

### Process Registration
```go
func (s *Server) RegisterProcessType(name string, factory ProcessFactory) {
    s.processes.Register(name, factory)
}
```

### Process Factory Interface
```go
type ProcessFactory interface {
    Create(config map[string]interface{}) (Process, error)
}
```

## Best Practices

### Thread Safety
- Use mutex protection for shared state
- RLock for reads, Lock for writes
- Keep lock durations minimal
- Copy data for responses

### Error Handling
- Validate inputs early
- Return appropriate errors
- Clean up on failure
- Log errors appropriately

### Resource Management
- Clean shutdown on context cancellation
- Proper file handling
- Memory management
- Connection handling