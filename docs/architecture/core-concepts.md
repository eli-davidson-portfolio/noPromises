# Core Concepts: Flow-Based Programming in Go

This document outlines the fundamental architectural concepts of our Flow-Based Programming (FBP) implementation in Go. These concepts strictly adhere to J. Paul Morrison's classical FBP principles while leveraging Go's native strengths in concurrent programming.

## 1. Foundational Elements

### Information Packets (IPs)

IPs are the fundamental unit of data flowing through the network. In our implementation, they are type-safe and support metadata:

```go
type IP[T any] struct {
    Type     IPType
    Data     T
    Metadata map[string]any
    Origin   string    // Source tracking
    FlowID   string    // Flow tracking
}

type IPType int

const (
    NormalIP IPType = iota
    InitialIP      // IIP
    OpenBracket
    CloseBracket
)
```

### Ports

Ports are the connection points through which processes send and receive IPs:

```go
type Port[T any] struct {
    Name        string
    Description string
    Required    bool
    Channel     chan *IP[T]
    MaxInputs   int  // For many-to-one connections
    MaxOutputs  int  // For one-to-many connections
}
```

### Processes

Processes are independent components that transform data. Each process runs in its own goroutine and communicates only through ports:

```go
type Process[In, Out any] interface {
    // Initialize is called once before processing begins
    Initialize(ctx context.Context) error
    
    // Process handles the main processing logic
    Process(ctx context.Context) error
    
    // Port management
    InPorts() []Port[In]
    OutPorts() []Port[Out]
    
    // InitialValues returns IIPs for configuration
    InitialValues() map[string]interface{}
    
    // Shutdown is called when processing ends
    Shutdown(ctx context.Context) error
}
```

## 2. Network Definition and Management

### Network

The Network is the core orchestrator that manages processes and their connections:

```go
type Network struct {
    nodes       map[string]interface{}
    connections map[string][]Connection
    iips        map[string]*IP[interface{}]
}
```

Key responsibilities:
- Process lifecycle management
- Connection establishment
- Network validation
- Error propagation
- Resource management

### Connections

Connections are typed channels that carry IPs between processes:

```go
type Connection struct {
    buffer    chan IP
    capacity  int
    metrics   *ConnectionMetrics
}
```

Features:
- Fixed buffer capacity
- Automatic backpressure handling
- Metrics collection
- Type safety enforcement

## 3. Error Handling and Recovery

### Error Propagation

Errors are handled at multiple levels:

```go
type NodeError struct {
    NodeID    string
    Severity  ErrorSeverity
    Err       error
    Timestamp time.Time
    Context   map[string]any
}

type ErrorSeverity int

const (
    SeverityDebug ErrorSeverity = iota
    SeverityInfo
    SeverityWarning
    SeverityError
    SeverityFatal
)
```

### Circuit Breaker

Protection against cascading failures:

```go
type CircuitBreaker struct {
    failures      int
    maxFailures   int
    resetTimeout  time.Duration
    lastFailure   time.Time
    state         circuitState
}
```

## 4. Resource Management

### Resource Pooling

Managed access to external resources:

```go
type ResourcePool[T Resource] struct {
    resources chan T
    factory   func() (T, error)
    size      int
    active    map[T]time.Time
}
```

### Resource Lifecycle

All resources implement a standard interface:

```go
type Resource interface {
    Initialize(ctx context.Context) error
    Close(ctx context.Context) error
    HealthCheck(ctx context.Context) error
}
```

## 5. Monitoring and Observability

### Metrics Collection

Comprehensive metrics for processes and connections:

```go
type Metrics struct {
    ProcessingTime  prometheus.Histogram
    MessageCount    prometheus.Counter
    ErrorCount      prometheus.Counter
    BufferUsage     prometheus.Gauge
}
```

### Health Checking

Regular health status monitoring:

```go
type HealthStatus struct {
    Status    string
    Details   map[string]string
    Timestamp time.Time
}
```

## 6. Bracket Handling

Support for hierarchical data structures:

```go
type BracketTracker struct {
    depth     int
    mu        sync.RWMutex
    onClose   func()
}
```

Features:
- Nested structure support
- Automatic bracket matching
- Substream processing

## 7. Key Design Principles

### Process Independence
- Each process runs in its own goroutine
- No shared state between processes
- Communication only through ports
- Clear ownership semantics

### Type Safety
- Generic type constraints
- Compile-time connection validation
- Interface-based contracts
- Full type inference

### Resource Safety
- Automatic cleanup
- Connection management
- Buffer control
- Lifecycle tracking

### Error Philosophy
- Process-level recovery
- Network error propagation
- Circuit breaking
- Graceful degradation

## 8. Best Practices

### Process Development
- Keep processes focused on single responsibility
- Use clear, descriptive port names
- Handle context cancellation properly
- Implement proper resource cleanup

### Network Design
- Use appropriate buffer sizes
- Monitor for potential deadlocks
- Implement error recovery strategies
- Document network topology

### Testing Requirements
- Process isolation tests
- Network integration tests
- Performance benchmarks
- Race condition verification

### Performance Considerations
- Explicit buffer sizing
- Resource pooling
- Monitoring hooks
- Backpressure handling