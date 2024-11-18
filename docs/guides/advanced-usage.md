# Advanced Usage Guide for noPromises

This guide covers advanced features and patterns in noPromises for building sophisticated Flow-Based Programming applications.

## Table of Contents
1. [Bracket Handling and Substreams](#bracket-handling-and-substreams)
2. [Advanced Error Handling](#advanced-error-handling)
3. [Resource Management](#resource-management)
4. [Dynamic Network Modification](#dynamic-network-modification)
5. [Performance Optimization](#performance-optimization)
6. [Advanced Testing Patterns](#advanced-testing-patterns)

## Bracket Handling and Substreams

Brackets enable hierarchical data processing in FBP networks. Here's how to implement bracket-aware processes:

```go
// BracketAwareProcess handles nested data structures
type BracketAwareProcess[T any] struct {
    inPort     *nop.Port[T]
    outPort    *nop.Port[T]
    bracketMgr *nop.BracketManager[T]
    depth      int32
}

func NewBracketAwareProcess[T any]() *BracketAwareProcess[T] {
    p := &BracketAwareProcess[T]{
        inPort:  nop.NewPort[T]("in", true),
        outPort: nop.NewPort[T]("out", true),
    }
    p.bracketMgr = nop.NewBracketManager(func() {
        // Called when a substream is complete
        p.processSubstream()
    })
    return p
}

func (p *BracketAwareProcess[T]) Process(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case ip := <-p.inPort.Channel:
            switch ip.Type {
            case nop.OpenBracket:
                p.bracketMgr.OpenBracket()
            case nop.CloseBracket:
                p.bracketMgr.CloseBracket()
            case nop.NormalIP:
                p.processIP(ip)
            }
        }
    }
}
```

### Processing Substreams

Here's how to process complete substreams:

```go
type SubstreamProcessor[T any] struct {
    buffer []T
    processor func([]T) []T
}

func (p *SubstreamProcessor[T]) ProcessSubstream(substream []T) []T {
    // Process entire substream as a unit
    result := p.processor(substream)
    
    // Send with brackets
    p.outPort.Channel <- &nop.IP[T]{Type: nop.OpenBracket}
    for _, item := range result {
        p.outPort.Channel <- &nop.IP[T]{
            Type: nop.NormalIP,
            Data: item,
        }
    }
    p.outPort.Channel <- &nop.IP[T]{Type: nop.CloseBracket}
    return result
}
```

## Advanced Error Handling

Implement sophisticated error handling with circuit breakers and recovery strategies:

```go
type CircuitBreaker struct {
    failures      int
    maxFailures   int
    resetTimeout  time.Duration
    lastFailure   time.Time
    state         circuitState
    mu            sync.RWMutex
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:       stateClosed,
    }
}

// Example usage in a process
type ResilientProcess[In, Out any] struct {
    *BaseProcess[In, Out]
    circuitBreaker *CircuitBreaker
}

func (p *ResilientProcess[In, Out]) Process(ctx context.Context) error {
    return p.circuitBreaker.Execute(func() error {
        return p.BaseProcess.Process(ctx)
    })
}
```

### Error Propagation

Implement network-wide error handling:

```go
type ErrorHandler struct {
    handlers   map[nop.ErrorSeverity][]func(*nop.NodeError)
    errorChan  chan *nop.NodeError
}

func (eh *ErrorHandler) PropagateError(err *nop.NodeError) {
    // Determine error severity and action
    switch err.Severity {
    case nop.SeverityFatal:
        eh.handleFatalError(err)
    case nop.SeverityError:
        eh.handleRecoverableError(err)
    case nop.SeverityWarning:
        eh.handleWarning(err)
    }
}
```

## Resource Management

Implement proper resource pooling and lifecycle management:

```go
type ResourcePool[T nop.Resource] struct {
    resources chan T
    factory   func() (T, error)
    size      int
}

func (p *ResourcePool[T]) Acquire(ctx context.Context) (T, error) {
    select {
    case resource := <-p.resources:
        if err := resource.HealthCheck(ctx); err != nil {
            // Resource is unhealthy, create new one
            return p.factory()
        }
        return resource, nil
    case <-ctx.Done():
        var zero T
        return zero, ctx.Err()
    }
}

// Example database connection pool
type DBPool struct {
    *ResourcePool[*sql.DB]
}

func NewDBPool(size int, connStr string) *DBPool {
    return &DBPool{
        ResourcePool: NewResourcePool(size, func() (*sql.DB, error) {
            return sql.Open("postgres", connStr)
        }),
    }
}
```

## Dynamic Network Modification

Modify networks at runtime:

```go
type DynamicNetwork struct {
    *nop.Network
    mu sync.RWMutex
}

func (n *DynamicNetwork) AddProcessDynamic(name string, p nop.Process) error {
    n.mu.Lock()
    defer n.mu.Unlock()
    
    // Validate process can be added
    if err := n.validateProcessAddition(name, p); err != nil {
        return err
    }
    
    // Add process and reconnect affected connections
    return n.addAndReconnect(name, p)
}

func (n *DynamicNetwork) SwapProcess(name string, newP nop.Process) error {
    n.mu.Lock()
    defer n.mu.Unlock()
    
    oldP := n.GetProcess(name)
    // Gracefully drain old process
    oldP.Drain()
    // Hot-swap to new process
    return n.replaceProcess(name, newP)
}
```

## Performance Optimization

Implement advanced performance optimizations:

```go
type BufferSizer struct {
    metrics    []nop.PerformanceMetrics
    window     time.Duration
    targetLoad float64
}

func (bs *BufferSizer) CalculateOptimalSize() int {
    // Use metrics to calculate optimal buffer size
    throughput := bs.calculateAverageThroughput()
    latency := bs.calculateAverageLatency()
    return int(throughput * latency.Seconds() * bs.targetLoad)
}

// Implement zero-copy where possible
type ZeroCopyTransform[T any] struct {
    transform func(*T)
}

func (z *ZeroCopyTransform[T]) Process(ip *nop.IP[T]) {
    // Modify data in place when possible
    z.transform(&ip.Data)
}
```

## Advanced Testing Patterns

Implement sophisticated testing strategies:

```go
type NetworkTester struct {
    network    *nop.Network
    harnesses  map[string]interface{}
    mocks      map[string]*MockProcess
    collector  *nop.MetricsCollector
}

func (nt *NetworkTester) SimulateLoad(ctx context.Context, rate float64) error {
    ticker := time.NewTicker(time.Duration(1/rate * float64(time.Second)))
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := nt.sendTestMessage(); err != nil {
                return err
            }
        }
    }
}

// Property-based testing
func TestNetworkProperties(t *testing.T) {
    network := setupTestNetwork(t)
    
    rapid.Check(t, func(t *rapid.T) {
        // Generate random valid inputs
        input := rapid.SliceOf(rapid.String()).Draw(t, "input")
        
        // Process through network
        output := processThrough(network, input)
        
        // Verify properties
        assert.Equal(t, len(input), len(output))
        assert.True(t, maintainsOrder(input, output))
    })
}
```

## Advanced Configuration

Use type-safe configuration with validation:

```go
type ProcessConfig[T any] struct {
    ID          string          `json:"id"`
    Type        string          `json:"type"`
    Config      T              `json:"config"`
    Properties  map[string]any `json:"properties,omitempty"`
}

func (c *ProcessConfig[T]) Validate() error {
    if c.ID == "" {
        return errors.New("id is required")
    }
    
    if c.Type == "" {
        return errors.New("type is required")
    }
    
    if validator, ok := any(c.Config).(interface{ Validate() error }); ok {
        return validator.Validate()
    }
    
    return nil
}
```

## Performance Tips

1. **Buffer Sizing**
   - Use metrics to determine optimal sizes
   - Monitor and adjust dynamically
   - Consider memory constraints

2. **Resource Pooling**
   - Pool expensive resources
   - Implement proper cleanup
   - Use health checks

3. **Memory Management**
   - Implement zero-copy where possible
   - Use object pools for IPs
   - Monitor GC pressure

4. **Concurrency**
   - Balance parallelism
   - Use appropriate channel sizes
   - Implement backpressure

Remember: With great power comes great responsibility. These advanced features should be used judiciously and only when their complexity is warranted by your use case.