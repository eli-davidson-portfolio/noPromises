# Monitoring and Observability Architecture

This document details the monitoring and observability system in our FBP implementation, which provides comprehensive insights into network operation, component performance, and system health.

## Core Components

### Metrics Collector

```go
type MetricsCollector struct {
    meter          metric.Meter
    tracer         trace.Tracer
    nodeMetrics    map[string]*NodeMetrics
    networkMetrics *NetworkMetrics
    mu            sync.RWMutex
}

func NewMetricsCollector(serviceName string) (*MetricsCollector, error) {
    meter := otel.GetMeterProvider().Meter(
        "goflows",
        metric.WithInstrumentationVersion("1.0.0"),
    )
    
    tracer := otel.GetTracerProvider().Tracer(
        "goflows",
        trace.WithInstrumentationVersion("1.0.0"),
    )
    
    return &MetricsCollector{
        meter:       meter,
        tracer:      tracer,
        nodeMetrics: make(map[string]*NodeMetrics),
        networkMetrics: &NetworkMetrics{
            connections: make(map[string]*ConnectionMetrics),
        },
    }, nil
}
```

### Node Metrics

```go
type NodeMetrics struct {
    // Counters
    messagesProcessed metric.Int64Counter
    errorsTotal      metric.Int64Counter
    
    // Histograms
    processingTime   metric.Float64Histogram
    queueLength     metric.Int64Histogram
    
    // Gauges
    currentLoad     metric.Float64UpDownCounter
    memoryUsage     metric.Int64UpDownCounter
    
    // Latest values
    lastProcessed   time.Time
    currentStatus   NodeStatus
    mu             sync.RWMutex
}
```

### Network Metrics

```go
type NetworkMetrics struct {
    // Counters
    totalMessages   metric.Int64Counter
    totalErrors     metric.Int64Counter
    
    // Histograms
    flowLatency     metric.Float64Histogram
    
    // Gauges
    activeNodes     metric.Int64UpDownCounter
    totalMemory     metric.Int64UpDownCounter
    
    // Connection metrics
    connections    map[string]*ConnectionMetrics
}
```

### Connection Metrics

```go
type ConnectionMetrics struct {
    bufferUsage    metric.Float64Histogram
    throughput     metric.Float64Histogram
    backpressure   metric.Int64Counter
}
```

## Monitored Components

### Monitored Node

```go
type MonitoredNode[In, Out any] struct {
    Node[In, Out]
    metrics    *NodeMetrics
    collector  *MetricsCollector
    nodeID     string
}

func (n *MonitoredNode[In, Out]) Process(ctx context.Context, in Port[In], out Port[Out]) error {
    // Create span for tracing
    ctx, span := n.collector.tracer.Start(ctx, 
        fmt.Sprintf("node.%s.process", n.nodeID),
        trace.WithAttributes(attribute.String("node.id", n.nodeID)),
    )
    defer span.End()
    
    // Track processing time
    start := time.Now()
    defer func() {
        duration := time.Since(start).Seconds()
        n.metrics.processingTime.Record(ctx, duration)
    }()
    
    // Process messages with monitoring
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
            
        case msg, ok := <-in:
            if !ok {
                return nil
            }
            
            // Create span for message processing
            msgCtx, msgSpan := n.collector.tracer.Start(ctx, 
                fmt.Sprintf("node.%s.process_message", n.nodeID))
            
            // Process message
            err := n.processMessage(msgCtx, msg, out)
            
            // Record metrics
            n.metrics.messagesProcessed.Add(ctx, 1)
            
            if err != nil {
                msgSpan.RecordError(err)
                msgSpan.End()
                return err
            }
            
            msgSpan.End()
        }
    }
}
```

### Monitored Connection

```go
type MonitoredConnection[T any] struct {
    Connection[T]
    metrics *ConnectionMetrics
}

func (c *MonitoredConnection[T]) Send(ctx context.Context, value T) error {
    start := time.Now()
    
    err := c.Connection.Send(ctx, value)
    
    // Record metrics
    duration := time.Since(start).Seconds()
    c.metrics.throughput.Record(ctx, duration)
    
    if err != nil {
        // Record backpressure if send failed due to full buffer
        c.metrics.backpressure.Add(ctx, 1)
        return err
    }
    
    return nil
}
```

## Health Checking System

### Health Checker

```go
type HealthChecker struct {
    collector *MetricsCollector
    checks    map[string]HealthCheck
    mu        sync.RWMutex
}

type HealthCheck struct {
    Check    func(context.Context) error
    Timeout  time.Duration
    Critical bool
}

type HealthStatus struct {
    Status    string
    Details   map[string]string
    Timestamp time.Time
}

func (hc *HealthChecker) RunChecks(ctx context.Context) *HealthStatus {
    status := &HealthStatus{
        Status:    "healthy",
        Details:   make(map[string]string),
        Timestamp: time.Now(),
    }
    
    hc.mu.RLock()
    defer hc.mu.RUnlock()
    
    for name, check := range hc.checks {
        checkCtx, cancel := context.WithTimeout(ctx, check.Timeout)
        err := check.Check(checkCtx)
        cancel()
        
        if err != nil {
            status.Details[name] = err.Error()
            if check.Critical {
                status.Status = "unhealthy"
            }
        } else {
            status.Details[name] = "ok"
        }
    }
    
    return status
}
```

## Visualization Components

```tsx
const NetworkVisualizer = ({ network, metrics, className = '' }) => {
    const [selectedNode, setSelectedNode] = useState(null);
    const [timeRange, setTimeRange] = useState('5m');
    
    return (
        <div className={`flex flex-col space-y-4 p-4 ${className}`}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <NetworkGraph 
                    network={network} 
                    onNodeSelect={setSelectedNode}
                    metrics={metrics}
                />
                {selectedNode && (
                    <NodeDetails 
                        node={selectedNode}
                        metrics={metrics}
                        timeRange={timeRange}
                    />
                )}
            </div>
            <NetworkMetrics 
                metrics={metrics}
                timeRange={timeRange}
                onTimeRangeChange={setTimeRange}
            />
            <NetworkAlerts network={network} metrics={metrics} />
        </div>
    );
};
```

## Best Practices

1. **Metric Collection**
   - Use appropriate metric types
   - Include relevant labels
   - Set meaningful bucket sizes
   - Monitor memory usage

2. **Tracing**
   - Create spans for key operations
   - Add relevant attributes
   - Track error contexts
   - Maintain trace context

3. **Health Checks**
   - Define critical checks
   - Set appropriate timeouts
   - Handle check failures
   - Monitor check timing

4. **Resource Monitoring**
   - Track resource usage
   - Monitor pool utilization
   - Watch for leaks
   - Set alerts for thresholds

## Alert Configuration

```go
// Example alert configuration
const alerts = {
    highErrorRate: {
        condition: (metrics) => metrics.errorRate > 10,
        severity: "error",
        message: "High error rate detected",
    },
    backpressure: {
        condition: (metrics) => metrics.backpressure > 0,
        severity: "warning",
        message: "Backpressure detected",
    },
    resourceExhaustion: {
        condition: (metrics) => metrics.resourceUsage > 90,
        severity: "critical",
        message: "Resource pool near exhaustion",
    },
}
```

## Usage Examples

### Basic Monitoring Setup

```go
// Create metrics collector
collector, err := NewMetricsCollector("example-service")
if err != nil {
    panic(err)
}

// Create monitored node
baseNode := &SomeNode{}
monitoredNode, err := NewMonitoredNode("node1", baseNode, collector)
if err != nil {
    panic(err)
}

// Run node with monitoring
ctx := context.Background()
go func() {
    if err := monitoredNode.Process(ctx, in, out); err != nil {
        fmt.Printf("Node error: %v\n", err)
    }
}()
```

### Health Check Configuration

```go
checker := NewHealthChecker(collector)

// Add checks
checker.AddCheck("database", HealthCheck{
    Check: checkDatabase,
    Timeout: time.Second * 5,
    Critical: true,
})

checker.AddCheck("cache", HealthCheck{
    Check: checkCache,
    Timeout: time.Second * 2,
    Critical: false,
})
```

## Monitoring Dashboard Integration

The monitoring system exports metrics in a format compatible with common monitoring systems:

- Prometheus metrics endpoint
- OpenTelemetry trace export
- Health check API endpoint
- Real-time visualization updates

These integrations enable comprehensive monitoring and alerting through standard tools and practices.