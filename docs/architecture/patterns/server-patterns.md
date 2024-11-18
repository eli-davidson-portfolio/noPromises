# Server Design Patterns

This document outlines key patterns for implementing the Flow Server subsystem.

## Network Management Patterns

### Network Lifecycle Management
```go
type NetworkManager struct {
    networks map[string]*ManagedNetwork
    mu       sync.RWMutex
}

type ManagedNetwork struct {
    Network   *network.Network
    Status    NetworkStatus
    StartTime time.Time
    Config    FlowConfig
}
```

### Safe Network Operations
```go
func (m *NetworkManager) SafeOperation(id string, op func(*ManagedNetwork) error) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    network, exists := m.networks[id]
    if !exists {
        return ErrNetworkNotFound
    }
    
    return op(network)
}
```

### Network State Transitions
```go
type NetworkStatus string

const (
    NetworkStatusCreated   NetworkStatus = "created"
    NetworkStatusStarting  NetworkStatus = "starting"
    NetworkStatus Running  NetworkStatus = "running"
    NetworkStatusStopping NetworkStatus = "stopping"
    NetworkStatusStopped  NetworkStatus = "stopped"
    NetworkStatusError    NetworkStatus = "error"
)
```

## Process Registry Patterns

### Factory Registration
```go
func (r *ProcessRegistry) Register(name string, factory ProcessFactory) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.processes[name]; exists {
        return ErrProcessTypeExists
    }
    
    r.processes[name] = factory
    return nil
}
```

### Process Configuration Validation
```go
type ProcessConfig struct {
    Type       string                 `json:"type"`
    Config     map[string]interface{} `json:"config"`
    Validation func() error
}

func (c *ProcessConfig) Validate() error {
    if c.Type == "" {
        return ErrMissingProcessType
    }
    if c.Validation != nil {
        return c.Validation()
    }
    return nil
}
```

## Request Handling Patterns

### Request Validation
```go
func validateFlowConfig(config FlowConfig) error {
    if config.ID == "" {
        return ErrMissingFlowID
    }
    
    if len(config.Nodes) == 0 {
        return ErrNoNodes
    }
    
    return validateConnections(config.Edges)
}
```

### Response Formatting
```go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func NewSuccessResponse(data interface{}) APIResponse {
    return APIResponse{
        Success: true,
        Data:    data,
    }
}
```

## Error Handling Patterns

### Error Types
```go
var (
    ErrNetworkNotFound   = errors.New("network not found")
    ErrNetworkExists     = errors.New("network already exists")
    ErrInvalidState      = errors.New("invalid network state")
    ErrProcessTypeExists = errors.New("process type already registered")
)
```

### Error Recovery
```go
func (s *FlowServer) handleCreateFlow(w http.ResponseWriter, r *http.Request) {
    var config FlowConfig
    if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
        s.sendError(w, http.StatusBadRequest, err)
        return
    }
    
    if err := s.createFlow(r.Context(), config); err != nil {
        switch {
        case errors.Is(err, ErrNetworkExists):
            s.sendError(w, http.StatusConflict, err)
        case errors.Is(err, ErrInvalidConfig):
            s.sendError(w, http.StatusBadRequest, err)
        default:
            s.sendError(w, http.StatusInternalServerError, err)
        }
        return
    }
    
    s.sendResponse(w, http.StatusCreated, NewSuccessResponse(config))
}
```

## Concurrency Patterns

### Safe State Access
```go
func (s *FlowServer) GetNetworkStatus(id string) (NetworkStatus, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    network, exists := s.networks[id]
    if !exists {
        return "", ErrNetworkNotFound
    }
    
    return network.Status, nil
}
```

### Background Tasks
```go
func (s *FlowServer) startMonitoring(ctx context.Context) {
    ticker := time.NewTicker(monitorInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            s.checkNetworks()
        }
    }
}
```

## Best Practices

### Request Handling
- Validate all input
- Use appropriate HTTP status codes
- Provide detailed error messages
- Handle timeouts

### State Management
- Use mutex protection
- Atomic state transitions
- Safe concurrent access
- Clean resource cleanup

### Error Handling
- Proper error types
- Consistent error responses
- Recovery procedures
- Error logging

### Testing
- Test concurrent operations
- Verify state transitions
- Check error conditions
- Test cleanup procedures 