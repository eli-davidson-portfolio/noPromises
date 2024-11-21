# Metrics Collection Architecture

## Core Metrics Types

### Process Metrics
```go
type ProcessMetrics struct {
    // Message flow metrics
    MessagesProcessed *prometheus.CounterVec
    MessageLatency    *prometheus.HistogramVec
    QueueDepth        *prometheus.GaugeVec
    
    // Resource metrics
    ProcessingTime    *prometheus.HistogramVec
    MemoryUsage      *prometheus.GaugeVec
    GoroutineCount   prometheus.Gauge
    
    // Error metrics
    ErrorCount       *prometheus.CounterVec
    LastError        *prometheus.GaugeVec
}

func NewProcessMetrics(processID string) *ProcessMetrics {
    return &ProcessMetrics{
        MessagesProcessed: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "process_messages_total",
                Help: "Total number of messages processed",
            },
            []string{"process_id", "port"},
        ),
        // ... other metric initializations
    }
}
```

### Network Metrics
```go
type NetworkMetrics struct {
    // Topology metrics
    ActiveProcesses   prometheus.Gauge
    ActiveConnections prometheus.Gauge
    
    // Flow metrics
    NetworkLatency    *prometheus.HistogramVec
    MessageBacklog    *prometheus.GaugeVec
    
    // Health metrics
    CircuitBreakers   *prometheus.GaugeVec
    PartitionCount    prometheus.Gauge
}
```

## Collection Mechanisms

### Metric Collection
```go
type MetricsCollector struct {
    registry    *prometheus.Registry
    processes   map[string]*ProcessMetrics
    network     *NetworkMetrics
    mu          sync.RWMutex
}

func (mc *MetricsCollector) RecordMessage(processID, port string, latency time.Duration) {
    mc.mu.RLock()
    if pm, exists := mc.processes[processID]; exists {
        pm.MessagesProcessed.WithLabelValues(processID, port).Inc()
        pm.MessageLatency.WithLabelValues(processID, port).Observe(latency.Seconds())
    }
    mc.mu.RUnlock()
}
```

### Resource Monitoring
```go
type ResourceMonitor struct {
    metrics    *ProcessMetrics
    interval   time.Duration
    done       chan struct{}
}

func (rm *ResourceMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(rm.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            rm.collectResourceMetrics()
        }
    }
}
```

## Storage and Aggregation

### Time Series Storage
```go
type MetricStorage struct {
    db        *tsdb.DB
    retention time.Duration
    buffer    *ring.Buffer
}

func (ms *MetricStorage) Store(metric prometheus.Metric) error {
    // Store metric in time series database
    return ms.db.Write(context.Background(), metric)
}
```

### Metric Aggregation
```go
type Aggregator struct {
    window    time.Duration
    metrics   map[string]*aggregatedMetric
    mu        sync.RWMutex
}

type aggregatedMetric struct {
    count     int64
    sum       float64
    min       float64
    max       float64
    timestamp time.Time
}
```

## Integration Points

### Prometheus Integration
```go
func SetupMetrics(router *mux.Router) *MetricsCollector {
    collector := NewMetricsCollector()
    
    // Register metrics endpoint
    router.Handle("/metrics", promhttp.HandlerFor(
        collector.registry,
        promhttp.HandlerOpts{},
    ))
    
    return collector
}
```

### Process Integration
```go
type MetricsMiddleware struct {
    collector *MetricsCollector
    processID string
}

func (m *MetricsMiddleware) WrapProcess(p Process) Process {
    return &metricsWrapper{
        process:   p,
        collector: m.collector,
        processID: m.processID,
    }
}
```

## Best Practices

### Metric Naming
- Use consistent naming conventions
- Include relevant labels
- Follow Prometheus standards
- Document metric meanings

### Collection Strategy
- Minimize collection overhead
- Use appropriate metric types
- Set reasonable intervals
- Handle collection errors

### Resource Usage
- Monitor collection impact
- Buffer metrics appropriately
- Set retention policies
- Clean up old metrics

### Integration
- Use standard exporters
- Implement health checks
- Support metric scraping
- Enable metric discovery

### Testing
- Test metric recording
- Verify aggregations
- Check metric labels
- Test collection errors