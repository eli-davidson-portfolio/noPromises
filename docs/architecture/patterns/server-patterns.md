# Server Design Patterns

This document outlines key patterns implemented in the Flow Server subsystem.

## Core Components

### Server Configuration
```go
type Config struct {
    Port     int    // HTTP server port
    DocsPath string // Documentation path
}
```

### Server Structure
```go
type Server struct {
    config    Config
    router    *mux.Router
    flows     *FlowManager
    docs      *docs.Server
    Handler   http.Handler
}
```

## Documentation Server Patterns

### HTML Wrapper Pattern
```go
func (s *Server) renderDocPage(w http.ResponseWriter, content string) {
    html := `<!DOCTYPE html>
    <html>
        <head>
            <title>Documentation</title>
            <link rel="stylesheet" href="/css/markdown.css">
        </head>
        <body>
            <div class="markdown-body">
                {{ .Content }}
            </div>
        </body>
    </html>`
    // Render template with content
}
```

### Route Setup Pattern
```go
func (s *Server) setupRoutes() {
    // Documentation routes
    s.router.PathPrefix("/docs/").Handler(s.docs)
    s.router.Handle("/api-docs", s.docs.SwaggerUI())
    s.router.Handle("/diagrams/", s.docs.Diagrams())

    // API routes
    s.router.PathPrefix("/api/").Handler(s.api)
}
```

### Middleware Chain
```go
func (s *Server) setupMiddleware() {
    s.router.Use(
        middleware.Logger,
        middleware.Recoverer,
        middleware.RequestID,
        middleware.RealIP,
    )
}
```

## Response Patterns

### JSON Response
```go
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "data": data,
    })
}
```

### Error Response
```go
func respondError(w http.ResponseWriter, status int, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": map[string]string{
            "message": err.Error(),
        },
    })
}
```

## Documentation Patterns

### Markdown Processing
```go
func (s *Server) processMarkdown(content []byte) string {
    // Convert Markdown to HTML
    html := markdown.ToHTML(content)
    
    // Add syntax highlighting
    html = highlight.Code(html)
    
    return html
}
```

### Diagram Generation
```go
func (s *Server) generateDiagram(networkID string) (string, error) {
    network, err := s.flows.Get(networkID)
    if err != nil {
        return "", err
    }
    
    return s.docs.GenerateMermaid(network)
}
```

## Best Practices

### Error Handling
- Use appropriate HTTP status codes
- Provide clear error messages
- Include error details when safe
- Log errors appropriately

### Request Processing
- Validate input early
- Use appropriate content types
- Handle timeouts
- Support graceful shutdown

### Documentation
- Keep docs close to code
- Update docs with changes
- Include examples
- Test documentation

### Testing
- Unit test all patterns
- Test error cases
- Verify documentation
- Check response formats

## Flow Management

### Flow Manager
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
```

### Flow States
```go
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

## Process Registry

### Registry Structure
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

### Response Helpers
```go
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.WriteHeader(status)
    if data != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "data": data,
        })
    }
}

func respondError(w http.ResponseWriter, status int, err error) {
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": map[string]interface{}{
            "message": err.Error(),
        },
    })
}
```

### Flow Validation
```go
func validateFlowConfig(config map[string]interface{}) error {
    if config["id"] == nil {
        return fmt.Errorf("missing flow id")
    }

    nodes, ok := config["nodes"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid nodes configuration")
    }

    // Validate node configurations...
    return nil
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

## Server Patterns

### Route Handling
- Add documentation server patterns
- Update route handling patterns

### File Serving
- Add file serving patterns

### HTML Wrapper
- Document HTML wrapper patterns