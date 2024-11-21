# Server Metrics Architecture

## Core Server Metrics

### HTTP Server Metrics
```go
type HTTPServerMetrics struct {
    // Request metrics
    RequestCount    *prometheus.CounterVec
    RequestLatency  *prometheus.HistogramVec
    RequestSize     *prometheus.HistogramVec
    ResponseSize    *prometheus.HistogramVec
    
    // Status metrics
    ResponseStatus  *prometheus.CounterVec
    ActiveRequests  prometheus.Gauge
    
    // Connection metrics
    OpenConnections prometheus.Gauge
    ConnectionAge   *prometheus.HistogramVec
}

func NewHTTPServerMetrics() *HTTPServerMetrics {
    return &HTTPServerMetrics{
        RequestCount: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_requests_total",
                Help: "Total number of HTTP requests",
            },
            []string{"method", "path", "status"},
        ),
        // ... other metric initializations
    }
}
```

### Resource Metrics
```go
type ServerResourceMetrics struct {
    // System metrics
    CPUUsage        *prometheus.GaugeVec
    MemoryUsage     *prometheus.GaugeVec
    GoroutineCount  prometheus.Gauge
    ThreadCount     prometheus.Gauge
    
    // File descriptors
    OpenFiles       prometheus.Gauge
    MaxFiles        prometheus.Gauge
    
    // Network
    NetworkIO       *prometheus.CounterVec
    TCPConnections  *prometheus.GaugeVec
}
```

## Server Middleware

### Metrics Middleware
```go
type MetricsMiddleware struct {
    metrics *HTTPServerMetrics
}

func (m *MetricsMiddleware) Wrap(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Track active requests
        m.metrics.ActiveRequests.Inc()
        defer m.metrics.ActiveRequests.Dec()
        
        // Wrap response writer to capture status
        wrapped := NewResponseWriter(w)
        
        next.ServeHTTP(wrapped, r)
        
        // Record metrics
        duration := time.Since(start)
        m.metrics.RequestLatency.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(wrapped.Status()),
        ).Observe(duration.Seconds())
        
        m.metrics.RequestCount.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(wrapped.Status()),
        ).Inc()
    })
}
```

### Resource Tracking
```go
type ResourceTracker struct {
    metrics  *ServerResourceMetrics
    interval time.Duration
}

func (rt *ResourceTracker) Start(ctx context.Context) {
    ticker := time.NewTicker(rt.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            rt.collectResourceMetrics()
        }
    }
}
```

## Integration Points

### Flow Server Integration
```go
type FlowServerMetrics struct {
    // Flow management
    ActiveFlows     prometheus.Gauge
    FlowOperations  *prometheus.CounterVec
    
    // Process metrics
    ProcessStarts   *prometheus.CounterVec
    ProcessFailures *prometheus.CounterVec
    
    // Network metrics
    NetworkLatency  *prometheus.HistogramVec
    MessageDrops    *prometheus.CounterVec
}

func (s *Server) setupMetrics() {
    // Initialize server metrics
    s.metrics = NewFlowServerMetrics()
    
    // Add metrics middleware
    s.router.Use(NewMetricsMiddleware(s.metrics).Wrap)
    
    // Start resource tracking
    go s.resourceTracker.Start(s.ctx)
}
```

### Prometheus Integration
```go
func (s *Server) setupPrometheus() {
    // Create registry
    registry := prometheus.NewRegistry()
    
    // Register metric collectors
    registry.MustRegister(
        s.metrics.http,
        s.metrics.resources,
        s.metrics.flows,
    )
    
    // Add metrics endpoint
    s.router.Handle("/metrics", promhttp.HandlerFor(
        registry,
        promhttp.HandlerOpts{
            EnableOpenMetrics: true,
        },
    ))
}
```

## Best Practices

### Metric Design
- Use consistent naming
- Add appropriate labels
- Follow Prometheus conventions
- Document metric meanings

### Performance Impact
- Buffer metric updates
- Use appropriate metric types
- Monitor collection overhead
- Set reasonable intervals

### Resource Management
- Track system resources
- Monitor goroutines
- Watch file descriptors
- Check memory usage

### Integration
- Coordinate with monitoring
- Support metric aggregation
- Enable metric discovery
- Implement health checks

### Testing
- Test metric collection
- Verify metric values
- Check performance impact
- Test high load scenarios
