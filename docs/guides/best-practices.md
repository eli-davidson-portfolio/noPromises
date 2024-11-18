# Best Practices Guide for noPromises

This guide outlines proven patterns and practices for building robust, maintainable Flow-Based Programming applications with noPromises.

## 1. Process Design

### Single Responsibility
Each process should do one thing and do it well.

#### Bad ❌
```go
// Too many responsibilities
type SuperNode struct{}

func (n *SuperNode) Process(ctx context.Context, in, out Port[any]) error {
    for msg := range in {
        // Parse CSV
        data := parseCSV(msg)
        // Transform data
        transformed := transform(data)
        // Write to database
        writeDB(transformed)
        // Send notifications
        notify(transformed)
        out <- transformed
    }
    return nil
}
```

#### Good ✓
```go
// Single responsibility nodes
type CSVParser struct{}

func (p *CSVParser) Process(ctx context.Context, in Port[[]byte], out Port[[]Row]) error {
    for data := range in {
        rows := parseCSV(data)
        out <- rows
    }
    return nil
}

type DataTransformer[In, Out any] struct {
    Transform func(In) Out
}

func (t *DataTransformer[In, Out]) Process(ctx context.Context, in Port[In], out Port[Out]) error {
    for data := range in {
        result := t.Transform(data)
        out <- result
    }
    return nil
}
```

### Type Safety
Always use Go's type system to catch errors at compile time.

#### Bad ❌
```go
// Runtime type assertions
type UnsafeNode struct{}

func (n *UnsafeNode) Process(in, out chan interface{}) error {
    for data := range in {
        str, ok := data.(string)
        if !ok {
            return errors.New("expected string")
        }
        out <- strings.ToUpper(str)
    }
    return nil
}
```

#### Good ✓
```go
// Compile-time type safety
type SafeNode[In, Out any] struct {
    Process func(In) Out
}

func (n *SafeNode[In, Out]) Process(ctx context.Context, in Port[In], out Port[Out]) error {
    for data := range in {
        result := n.Process(data)
        out <- result
    }
    return nil
}

// Usage
uppercase := &SafeNode[string, string]{
    Process: strings.ToUpper,
}
```

### Port Management
Clear and explicit port definitions make networks easier to understand and maintain.

#### Bad ❌
```go
// Implicit ports
type Node struct {
    input  chan interface{}
    output chan interface{}
}
```

#### Good ✓
```go
// Explicit, documented ports
type Node[In, Out any] struct {
    InputPort  *Port[In] `port:"in,required"`
    OutputPort *Port[Out] `port:"out,required"`
    ErrorPort  *Port[error] `port:"error,optional"`
}
```

## 2. Error Handling

### Graceful Recovery
Implement comprehensive error handling and recovery strategies.

#### Bad ❌
```go
// No error handling
func (n *Node) Process(in, out chan Message) {
    for msg := range in {
        result := riskyOperation(msg)  // Might panic
        out <- result
    }
}
```

#### Good ✓
```go
func (n *Node) Process(ctx context.Context, in, out Port[Message]) error {
    defer func() {
        if r := recover(); r != nil {
            n.errorHandler.HandlePanic(r)
        }
    }()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg, ok := <-in:
            if !ok {
                return nil
            }
            if err := n.processMessage(msg, out); err != nil {
                n.errorHandler.HandleError(err)
                continue  // Keep processing other messages
            }
        }
    }
}
```

### Error Propagation
Define clear patterns for error handling across the network.

```go
type ErrorHandler struct {
    // Severity levels for different types of errors
    severityHandlers map[ErrorSeverity][]ErrorCallback
    
    // Channel for error events
    errorChan chan *ErrorEvent
    
    // Circuit breaker for preventing cascading failures
    circuitBreaker *CircuitBreaker
}
```

## 3. Resource Management

### Connection Pooling
Properly manage expensive resources.

```go
type ResourcePool[T Resource] struct {
    resources chan T
    factory   func() (T, error)
    
    // Monitor pool health
    metrics   *PoolMetrics
    
    // Lifecycle management
    ctx       context.Context
    cancel    context.CancelFunc
}

func (p *ResourcePool[T]) Acquire(ctx context.Context) (T, error) {
    select {
    case r := <-p.resources:
        if err := r.HealthCheck(); err != nil {
            // Resource is unhealthy, create new one
            return p.factory()
        }
        return r, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### Clean Shutdown
Ensure proper cleanup of resources.

```go
func (n *Network) Shutdown(ctx context.Context) error {
    n.cancel() // Cancel all processes
    
    // Wait for processes to finish
    done := make(chan struct{})
    go func() {
        n.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## 4. Testing

### Unit Testing Processes
Test processes in isolation.

```go
func TestTransformProcess(t *testing.T) {
    // Create test harness
    harness := NewNodeHarness[string, string](t)
    
    // Create process
    transform := NewTransformer(strings.ToUpper)
    
    // Test cases
    testCases := []struct{
        input    string
        expected string
    }{
        {"hello", "HELLO"},
        {"world", "WORLD"},
    }
    
    for _, tc := range testCases {
        result := harness.ProcessMessage(transform, tc.input)
        assert.Equal(t, tc.expected, result)
    }
}
```

### Network Testing
Test complete flows.

```go
func TestDataPipeline(t *testing.T) {
    // Create test network
    network := NewTestNetwork()
    
    // Add mock nodes
    network.AddMockSource("source", testData...)
    network.AddMockSink("sink")
    
    // Run network
    err := network.Run(context.Background())
    assert.NoError(t, err)
    
    // Verify results
    results := network.GetSinkData("sink")
    assert.Equal(t, expectedData, results)
}
```

## 5. Performance

### Monitoring
Implement comprehensive monitoring.

```go
type Metrics struct {
    ProcessingTime  prometheus.Histogram
    MessageCount    prometheus.Counter
    ErrorCount      prometheus.Counter
    BufferUsage     prometheus.Gauge
}

func (n *Node) Process(ctx context.Context, in, out Port[Message]) error {
    for msg := range in {
        start := time.Now()
        
        if err := n.processMessage(msg, out); err != nil {
            n.metrics.ErrorCount.Inc()
            continue
        }
        
        n.metrics.ProcessingTime.Observe(time.Since(start).Seconds())
        n.metrics.MessageCount.Inc()
    }
    return nil
}
```

### Backpressure
Handle backpressure gracefully.

```go
func (n *Node) Process(ctx context.Context, in, out Port[Message]) error {
    for msg := range in {
        select {
        case out <- msg:
            // Message sent successfully
        default:
            // Channel full - handle backpressure
            if err := n.handleBackpressure(ctx, msg, out); err != nil {
                return err
            }
        }
    }
    return nil
}
```

## 6. Documentation

### Self-Documenting Nodes
Make nodes self-documenting.

```go
type CustomerEnricher struct {
    // Document fields
    APIEndpoint string `doc:"External API endpoint for customer data"`
    CacheTime   time.Duration `doc:"How long to cache customer records"`
    
    Meta NodeMetadata `doc:"
        Enriches customer records with external data.
        
        Input: CustomerRecord
        Output: EnrichedCustomer
        
        Configuration:
        - APIEndpoint: External API endpoint
        - CacheTime: Cache duration
        
        Example:
            enricher := NewCustomerEnricher(
                WithEndpoint('https://api.example.com'),
                WithCache(time.Hour),
            )
    "`
}
```

## Key Takeaways

1. **Design for Isolation**
   - Keep processes independent
   - Use clear interfaces
   - Minimize shared state

2. **Embrace Type Safety**
   - Use generics
   - Define clear contracts
   - Catch errors at compile time

3. **Handle Failures Gracefully**
   - Implement error recovery
   - Use circuit breakers
   - Monitor health

4. **Test Thoroughly**
   - Unit test processes
   - Integration test networks
   - Benchmark performance

5. **Monitor Everything**
   - Track metrics
   - Watch for bottlenecks
   - Handle backpressure

Remember: Flow-Based Programming is about composing independent, reliable components. Follow these practices to build robust, maintainable systems.