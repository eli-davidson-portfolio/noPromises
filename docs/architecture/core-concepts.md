# Core Concepts: Flow-Based Programming in Go

This document outlines the fundamental architectural concepts of our Flow-Based Programming (FBP) implementation in Go.

## Server Components

### Server Configuration
```go
type Config struct {
    Port int
}
```

### Server Structure
```go
type Server struct {
    config    Config
    router    *mux.Router
    flows     *FlowManager
    processes *ProcessRegistry
    Handler   http.Handler
}
```

### Flow Management
```go
type FlowManager struct {
    flows map[string]*ManagedFlow
    mu    sync.RWMutex
}

type ManagedFlow struct {
    ID        string                 
    Config    map[string]interface{} 
    State     FlowState              
    StartTime *time.Time             
    Error     string                 
}

type FlowState string

const (
    FlowStateCreated  FlowState = "created"
    FlowStateStarting FlowState = "starting"
    FlowStateRunning  FlowState = "running"
    FlowStateStopping FlowState = "stopping"
    FlowStateStopped  FlowState = "stopped"
    FlowStateError    FlowState = "error"
)
```

### Process Registry
```go
type ProcessRegistry struct {
    processes map[string]ProcessFactory
    mu        sync.RWMutex
}

type ProcessFactory interface {
    Create(config map[string]interface{}) (Process, error)
}

type Process interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

## Request Handling

### Response Format
```go
// Success response
{
    "data": {
        // Response data
    }
}

// Error response
{
    "error": {
        "message": "Error description"
    }
}
```

### Flow Configuration
```json
{
    "id": "example-flow",
    "nodes": {
        "reader": {
            "type": "FileReader",
            "config": {
                "filename": "input.txt"
            }
        }
    },
    "edges": []
}
```

## Concurrency Patterns

### Safe State Access
```go
// Read access
s.flows.mu.RLock()
flow, exists := s.flows.flows[flowID]
s.flows.mu.RUnlock()

// Write access
s.flows.mu.Lock()
s.flows.flows[flowID] = flow
s.flows.mu.Unlock()
```

### Background Operations
```go
// Start flow in background
go func() {
    time.Sleep(50 * time.Millisecond)
    s.flows.mu.Lock()
    flow.State = FlowStateRunning
    s.flows.mu.Unlock()
}()
```

### Graceful Shutdown
```go
func (s *Server) Start(ctx context.Context) error {
    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", s.config.Port),
        Handler: s.Handler,
    }

    go func() {
        <-ctx.Done()
        srv.Shutdown(context.Background())
    }()

    return srv.ListenAndServe()
}
```

## Best Practices

### State Management
- Use mutex protection for shared state
- Prefer RLock for reads
- Keep lock durations minimal
- Copy data for responses when needed

### Error Handling
- Return appropriate HTTP status codes
- Provide clear error messages
- Log errors appropriately
- Clean up resources on error

### Request Processing
- Validate input early
- Use appropriate content types
- Handle timeouts
- Support graceful shutdown

### Testing
- Use table-driven tests
- Test concurrent operations
- Verify state transitions
- Check error conditions