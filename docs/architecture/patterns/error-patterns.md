# Error Handling Patterns

## Error Propagation Patterns

### Error Types
```go
// Base error type for flow errors
type FlowError struct {
    ProcessID string
    Op        string
    Err       error
    Context   map[string]interface{}
}

func (e *FlowError) Error() string {
    return fmt.Sprintf("%s: %s failed: %v", e.ProcessID, e.Op, e.Err)
}

func (e *FlowError) Unwrap() error {
    return e.Err
}
```

### Error Wrapping
```go
func (p *Process) Execute(ctx context.Context) error {
    if err := p.validate(); err != nil {
        return &FlowError{
            ProcessID: p.ID(),
            Op:       "validate",
            Err:      err,
            Context: map[string]interface{}{
                "state": p.state,
            },
        }
    }
    return nil
}
```

## Recovery Patterns

### Panic Recovery
```go
func WrapProcess(p Process) Process {
    return &recoveryWrapper{
        process: p,
        metrics: newMetrics(),
    }
}

type recoveryWrapper struct {
    process Process
    metrics *Metrics
}

func (w *recoveryWrapper) Process(ctx context.Context) error {
    defer func() {
        if r := recover(); r != nil {
            w.metrics.PanicCount.Inc()
            // Log stack trace
            debug.PrintStack()
        }
    }()
    return w.process.Process(ctx)
}
```

### Circuit Breaker Implementation
```go
type CircuitBreaker struct {
    failures     int
    maxFailures  int
    resetTimeout time.Duration
    lastFailure  time.Time
    mu           sync.RWMutex
}

func (cb *CircuitBreaker) Execute(op func() error) error {
    cb.mu.RLock()
    if cb.isOpen() {
        cb.mu.RUnlock()
        return ErrCircuitOpen
    }
    cb.mu.RUnlock()

    err := op()
    
    if err != nil {
        cb.mu.Lock()
        cb.recordFailure()
        cb.mu.Unlock()
    }
    
    return err
}
```

## Error Reporting Patterns

### Structured Error Logging
```go
type ErrorLogger struct {
    logger *zap.Logger
}

func (l *ErrorLogger) LogError(err error) {
    var flowErr *FlowError
    if errors.As(err, &flowErr) {
        l.logger.Error("flow error",
            zap.String("processID", flowErr.ProcessID),
            zap.String("operation", flowErr.Op),
            zap.Any("context", flowErr.Context),
            zap.Error(flowErr.Err),
        )
        return
    }
    l.logger.Error("unknown error", zap.Error(err))
}
```

### Error Metrics
```go
type ErrorMetrics struct {
    ErrorCount   *prometheus.CounterVec
    PanicCount   prometheus.Counter
    CircuitOpen  *prometheus.GaugeVec
}

func NewErrorMetrics() *ErrorMetrics {
    return &ErrorMetrics{
        ErrorCount: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "process_errors_total",
                Help: "Total number of process errors",
            },
            []string{"process_id", "error_type"},
        ),
        // ... other metric definitions
    }
}
```

## Best Practices

### Error Classification
- Use error types for categorization
- Wrap errors with context
- Preserve error chains
- Enable error inspection

### Recovery Strategy
- Recover only from panics
- Log recovery events
- Reset component state
- Report recovery metrics

### Error Handling
- Handle errors at appropriate levels
- Don't swallow errors
- Log with context
- Update component state

### Circuit Breaking
- Set appropriate thresholds
- Use exponential backoff
- Monitor breaker state
- Log state changes

### Monitoring
- Track error frequencies
- Monitor error patterns
- Alert on thresholds
- Analyze error context

### Testing
- Test error conditions
- Verify recovery
- Check error context
- Test circuit breakers