# Error Handling Architecture

This document describes the error handling architecture in our FBP implementation, based on classical FBP principles and Go's error handling patterns.

## Core Components

### Error Types

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

### Error Handler

The central error management component:

```go
type ErrorHandler struct {
    handlers   map[ErrorSeverity][]func(*NodeError)
    mu         sync.RWMutex
    errorChan  chan *NodeError
    ctx        context.Context
    cancel     context.CancelFunc
}
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
    mu            sync.RWMutex
}

type circuitState int

const (
    stateClosed circuitState = iota
    stateOpen
    stateHalfOpen
)
```

## Error Handling Process

1. **Error Detection**
   - Process-level errors captured through return values
   - Panics captured through deferred recovery
   - Resource errors detected through health checks

2. **Error Propagation**
   ```go
   func (eh *ErrorHandler) HandleError(err *NodeError) {
       select {
       case eh.errorChan <- err:
       default:
           // Error channel is full, log this somewhere
           fmt.Printf("Error channel full, dropping error: %v\n", err)
       }
   }
   ```

3. **Error Recovery**
   - Automatic retry for transient errors
   - Circuit breaker activation for persistent failures
   - Resource cleanup and reinitialization

4. **Network-Level Response**
   - Process isolation prevents cascade failures
   - Error propagation to dependent processes
   - Network state management during errors

## Circuit Breaker Implementation

```go
func (cb *CircuitBreaker) Execute(operation func() error) error {
    cb.mu.Lock()
    if cb.state == stateOpen {
        if time.Since(cb.lastFailure) > cb.resetTimeout {
            cb.state = stateHalfOpen
        } else {
            cb.mu.Unlock()
            return errors.New("circuit breaker is open")
        }
    }
    cb.mu.Unlock()

    err := operation()

    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()

        if cb.failures >= cb.maxFailures {
            cb.state = stateOpen
        }
        return err
    }

    if cb.state == stateHalfOpen {
        cb.state = stateClosed
    }
    cb.failures = 0
    return nil
}
```

## Error-Aware Components

### Error-Aware Node

```go
type ErrorAwareNode[In, Out any] struct {
    Node[In, Out]
    errorHandler *ErrorHandler
    circuitBreaker *CircuitBreaker
    nodeID      string
}

func (n *ErrorAwareNode[In, Out]) Process(ctx context.Context, in Port[In], out Port[Out]) error {
    return n.circuitBreaker.Execute(func() error {
        defer func() {
            if r := recover(); r != nil {
                err := &NodeError{
                    NodeID:    n.nodeID,
                    Severity:  SeverityFatal,
                    Err:       fmt.Errorf("panic: %v", r),
                    Timestamp: time.Now(),
                }
                n.errorHandler.HandleError(err)
            }
        }()

        return n.Node.Process(ctx, in, out)
    })
}
```

## Best Practices

1. **Error Categorization**
   - Use appropriate severity levels
   - Include context with errors
   - Track error patterns

2. **Circuit Breaker Configuration**
   - Set appropriate failure thresholds
   - Configure meaningful reset timeouts
   - Monitor breaker state

3. **Resource Management**
   - Always cleanup on errors
   - Implement proper shutdown
   - Handle partial failures

4. **Error Reporting**
   - Log all errors appropriately
   - Include relevant context
   - Track error frequencies

## Error Flow Examples

### Basic Error Flow
```go
func ExampleErrorFlow() {
    eh := NewErrorHandler()
    
    // Add error handlers
    eh.OnError(SeverityError, func(err *NodeError) {
        fmt.Printf("Error in node %s: %v\n", err.NodeID, err.Err)
    })

    eh.OnError(SeverityFatal, func(err *NodeError) {
        fmt.Printf("Fatal error in node %s: %v\n", err.NodeID, err.Err)
        // Initiate shutdown procedures
    })
}
```

### Circuit Breaker Usage
```go
func ExampleCircuitBreaker() {
    cb := NewCircuitBreaker(5, time.Minute)
    
    err := cb.Execute(func() error {
        // Attempt operation
        return someRiskyOperation()
    })
    
    if err != nil {
        // Handle error or circuit breaker open state
    }
}
```