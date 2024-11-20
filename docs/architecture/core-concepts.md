# Core Concepts

## Flow-Based Programming

noPromises implements Flow-Based Programming (FBP) principles in Go, providing a type-safe, concurrent execution environment.

### Key Components

#### Processes
```go
// Process interface defines core process behavior
type Process interface {
    Start(context.Context) error
    Stop(context.Context) error
    ID() string
}
```

#### Flows
```go
// Flow represents a network of connected processes
type Flow struct {
    ID     string
    State  string
    Config map[string]interface{}
}
```

#### Process Registry
```go
// ProcessFactory creates new process instances
type ProcessFactory interface {
    Create(config map[string]interface{}) (Process, error)
}
```

### Server Architecture

#### Core Server
```go
type Server struct {
    config    Config
    router    *mux.Router
    flows     *FlowManager
    processes *ProcessRegistry
    webServer *http.Server
}
```

#### Flow Management
```go
// FlowManager handles flow lifecycle
type FlowManager struct {
    flows map[string]*Flow
    mu    sync.RWMutex
}
```

### Documentation System

The server includes an integrated documentation system that provides:
- Markdown rendering
- API documentation
- Network visualization
- Live updates

### Type Safety

The implementation emphasizes type safety through:
- Strongly typed processes
- Type-safe message passing
- Configuration validation
- Error handling patterns
