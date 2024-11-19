# API Design Patterns

## Core Patterns

### Response Structure
All responses follow a consistent format:

```go
// Success response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.WriteHeader(status)
    if data != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "data": data,
        })
    }
}

// Error response
func respondError(w http.ResponseWriter, status int, err error) {
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": map[string]interface{}{
            "message": err.Error(),
        },
    })
}
```

### Middleware Chain
```go
func (s *Server) setupMiddleware() {
    s.router.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "application/json")
            next.ServeHTTP(w, r)
        })
    })
}
```

### Route Setup
```go
func (s *Server) setupRoutes() {
    api := s.router.PathPrefix("/api/v1").Subrouter()

    api.HandleFunc("/flows", s.handleCreateFlow).Methods(http.MethodPost)
    api.HandleFunc("/flows", s.handleListFlows).Methods(http.MethodGet)
    api.HandleFunc("/flows/{id}", s.handleGetFlow).Methods(http.MethodGet)
    api.HandleFunc("/flows/{id}", s.handleDeleteFlow).Methods(http.MethodDelete)
    api.HandleFunc("/flows/{id}/start", s.handleStartFlow).Methods(http.MethodPost)
    api.HandleFunc("/flows/{id}/stop", s.handleStopFlow).Methods(http.MethodPost)
    api.HandleFunc("/flows/{id}/status", s.handleGetFlowStatus).Methods(http.MethodGet)
}
```

### Request Validation
```go
func (s *Server) validateFlowConfig(config map[string]interface{}) error {
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

### Error Handling
Common HTTP status codes:
- 201: Resource created successfully
- 400: Invalid request or configuration
- 404: Resource not found
- 409: Resource state conflict
- 500: Server error

### Resource Management
```go
// Safe state access
s.flows.mu.RLock()
flow, exists := s.flows.flows[flowID]
s.flows.mu.RUnlock()

// State modification
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

## Testing Patterns

### Request Testing
```go
func (ts *TestServer) request(method, path string, body interface{}) *httptest.ResponseRecorder {
    var bodyReader *bytes.Buffer
    if body != nil {
        bodyBytes, err := json.Marshal(body)
        require.NoError(ts.t, err)
        bodyReader = bytes.NewBuffer(bodyBytes)
    } else {
        bodyReader = bytes.NewBuffer(nil)
    }

    req := httptest.NewRequest(method, path, bodyReader)
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    ts.Server.Handler.ServeHTTP(rr, req)
    return rr
}
```

### Table-Driven Tests
```go
tests := []struct {
    name       string
    flowConfig map[string]interface{}
    wantStatus int
    wantError  bool
}{
    {
        name: "valid flow",
        flowConfig: map[string]interface{}{
            "id": "test-flow",
            // ...
        },
        wantStatus: http.StatusCreated,
        wantError:  false,
    },
    // ...
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test implementation
    })
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