# Distributed Tracing Implementation

## Core Components

### Trace Context
```go
type TraceContext struct {
    TraceID     string
    SpanID      string
    ParentID    string
    ProcessID   string
    StartTime   time.Time
    Attributes  map[string]string
}

func NewTraceContext(processID string) *TraceContext {
    return &TraceContext{
        TraceID:    uuid.New().String(),
        SpanID:     uuid.New().String(),
        ProcessID:  processID,
        StartTime:  time.Now(),
        Attributes: make(map[string]string),
    }
}
```

### Span Management
```go
type Span struct {
    Context    *TraceContext
    Operation  string
    Status     SpanStatus
    Events     []SpanEvent
    StartTime  time.Time
    EndTime    time.Time
    mu         sync.RWMutex
}

type SpanEvent struct {
    Name      string
    Timestamp time.Time
    Attributes map[string]string
}

func (s *Span) AddEvent(name string, attrs map[string]string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.Events = append(s.Events, SpanEvent{
        Name:       name,
        Timestamp:  time.Now(),
        Attributes: attrs,
    })
}
```

## Trace Propagation

### Context Propagation
```go
type TracingPort[T any] struct {
    Port[T]
    tracer *Tracer
}

func (p *TracingPort[T]) Send(ctx context.Context, data T) error {
    span := p.tracer.StartSpan(ctx, "port.send")
    defer span.End()
    
    // Inject trace context into message
    msg := &TracedMessage[T]{
        Data:    data,
        Context: span.Context,
    }
    
    return p.Port.Send(ctx, msg.Data)
}
```

### Message Tracing
```go
type TracedMessage[T any] struct {
    Data    T
    Context *TraceContext
}

func (m *TracedMessage[T]) ExtractContext() *TraceContext {
    return m.Context
}
```

## Sampling Strategies

### Sampler Interface
```go
type Sampler interface {
    ShouldSample(traceID string) bool
    GetSamplingRate() float64
}

type RateLimitingSampler struct {
    rate      float64
    threshold uint64
    counter   atomic.Uint64
}

func (s *RateLimitingSampler) ShouldSample(traceID string) bool {
    count := s.counter.Add(1)
    return float64(count) <= float64(s.threshold)*s.rate
}
```

## Integration Systems

### OpenTelemetry Integration
```go
type OpenTelemetryExporter struct {
    client     *otel.Client
    processor  *BatchSpanProcessor
    attributes map[string]string
}

func (e *OpenTelemetryExporter) ExportSpan(span *Span) error {
    otelSpan := e.convertSpan(span)
    return e.processor.ProcessSpan(otelSpan)
}
```

### Process Integration
```go
type TracedProcess struct {
    Process
    tracer *Tracer
}

func (p *TracedProcess) Process(ctx context.Context) error {
    span := p.tracer.StartSpan(ctx, "process.execute")
    defer span.End()
    
    ctx = context.WithValue(ctx, tracingKey, span)
    return p.Process.Process(ctx)
}
```

## Performance Impact

### Buffer Management
```go
type SpanBuffer struct {
    spans    []*Span
    size     int
    interval time.Duration
    export   chan<- []*Span
    mu       sync.Mutex
}

func (b *SpanBuffer) Add(span *Span) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    b.spans = append(b.spans, span)
    if len(b.spans) >= b.size {
        b.flush()
    }
}
```

### Resource Management
```go
type TracerConfig struct {
    BufferSize      int
    FlushInterval   time.Duration
    MaxQueueSize    int
    WorkerCount     int
    SamplingRate    float64
}

func NewTracer(config TracerConfig) *Tracer {
    return &Tracer{
        buffer:    NewSpanBuffer(config.BufferSize),
        sampler:   NewRateLimitingSampler(config.SamplingRate),
        processor: NewBatchProcessor(config),
    }
}
```

## Best Practices

### Trace Context
- Propagate through all components
- Maintain parent-child relationships
- Include relevant attributes
- Handle context cancellation

### Sampling
- Use appropriate sampling rates
- Implement adaptive sampling
- Consider trace importance
- Monitor sampling impact

### Performance
- Buffer spans efficiently
- Batch export operations
- Monitor resource usage
- Handle backpressure

### Integration
- Support multiple exporters
- Handle export failures
- Maintain context consistency
- Clean up resources

### Testing
- Test trace propagation
- Verify sampling logic
- Check performance impact
- Test error conditions