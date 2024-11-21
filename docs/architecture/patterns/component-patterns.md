# Component Implementation Patterns

## Standard Component Structure

### Base Component
```go
type Component struct {
    id      string
    config  Config
    state   State
    ports   Ports
    metrics Metrics
}

type Ports struct {
    In  map[string]Port[any]
    Out map[string]Port[any]
}
```

### Component Lifecycle
```go
type Component interface {
    Initialize(ctx context.Context) error
    Process(ctx context.Context) error
    Shutdown(ctx context.Context) error
    Status() Status
}
```

## Reusable Patterns

### Port Management
```go
func (c *Component) setupPorts() error {
    // Input ports
    c.ports.In["data"] = ports.NewInput[string]("data", "Input data", true)
    
    // Output ports
    c.ports.Out["result"] = ports.NewOutput[string]("result", "Processed result", true)
    
    return nil
}
```

### Configuration Validation
```go
func (c *Component) validateConfig() error {
    if c.config.BufferSize < 1 {
        return fmt.Errorf("buffer size must be positive")
    }
    
    if c.config.Timeout < time.Second {
        return fmt.Errorf("timeout must be at least 1 second")
    }
    
    return nil
}
```

### Error Handling
```go
func (c *Component) Process(ctx context.Context) error {
    defer func() {
        if r := recover(); r != nil {
            c.metrics.PanicCount.Inc()
            c.state = StateError
        }
    }()
    
    // Process implementation
    return nil
}
```

## Anti-Patterns to Avoid

### Shared State
- Don't share memory between components
- Use message passing instead
- Maintain component isolation
- Avoid global state

### Resource Leaks
- Always clean up in Shutdown
- Use defer for cleanup
- Handle partial initialization
- Close all channels

### Error Handling
- Don't swallow errors
- Avoid panic in normal flow
- Don't return nil errors
- Handle all error cases

## Best Practices

### Component Design
- Single responsibility
- Clear interfaces
- Strong encapsulation
- Proper cleanup

### State Management
- Immutable where possible
- Protected mutable state
- Clear state transitions
- State validation

### Testing
- Unit test all components
- Test error conditions
- Verify cleanup
- Check metrics

### Documentation
- Document all ports
- Clear configuration
- Usage examples
- Error conditions