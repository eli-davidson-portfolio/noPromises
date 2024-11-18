# Performance Benchmarking Architecture

This document details the benchmarking architecture for measuring and optimizing performance in our FBP implementation.

## Core Components

### Benchmark Node

```go
type BenchmarkNode[In, Out any] struct {
    node      Node[In, Out]
    collector *MetricsCollector
}

func NewBenchmarkNode[In, Out any](node Node[In, Out]) *BenchmarkNode[In, Out] {
    collector, _ := NewMetricsCollector("benchmark")
    return &BenchmarkNode[In, Out]{
        node:      node,
        collector: collector,
    }
}
```

### Performance Metrics

```go
type PerformanceMetrics struct {
    Throughput      float64
    Latency        time.Duration
    ErrorRate      float64
    ResourceUsage  ResourceStats
}

type ResourceStats struct {
    MemoryUsage    uint64
    CPUUsage       float64
    GoroutineCount int
    BufferUsage    float64
}
```

### Benchmark Harness

```go
type BenchmarkHarness struct {
    node        Node
    inputPorts  map[string]chan IP
    outputPorts map[string]chan IP
    metrics     *PerformanceMetrics
    t           *testing.B
}

func NewBenchmarkHarness(b *testing.B, node Node) *BenchmarkHarness {
    return &BenchmarkHarness{
        node:       node,
        metrics:    &PerformanceMetrics{},
        t:          b,
    }
}
```

## Benchmarking Operations

### Running Benchmarks

```go
func (bn *BenchmarkNode[In, Out]) RunBenchmark(b *testing.B, input []In) {
    ctx := context.Background()
    in := make(Port[In], len(input))
    out := make(Port[Out], len(input))
    
    monitored, _ := NewMonitoredNode("benchmark", bn.node, bn.collector)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Send input
        for _, msg := range input {
            in <- msg
        }
        
        // Process messages
        if err := monitored.Process(ctx, in, out); err != nil {
            b.Fatal(err)
        }
        
        // Drain output
        for range input {
            <-out
        }
    }
}
```

### Measuring Resource Usage

```go
func (h *BenchmarkHarness) measureResources() ResourceStats {
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    
    return ResourceStats{
        MemoryUsage:    stats.Alloc,
        GoroutineCount: runtime.NumGoroutine(),
        BufferUsage:    h.calculateBufferUsage(),
    }
}

func (h *BenchmarkHarness) calculateBufferUsage() float64 {
    total := 0
    used := 0
    
    for _, port := range h.inputPorts {
        total += cap(port)
        used += len(port)
    }
    
    if total == 0 {
        return 0
    }
    return float64(used) / float64(total)
}
```

## Benchmark Patterns

### Throughput Benchmark

```go
func BenchmarkNodeThroughput(b *testing.B) {
    node := NewTestNode()
    benchmark := NewBenchmarkNode(node)
    
    input := generateTestData(1000)
    
    b.Run("throughput", func(b *testing.B) {
        benchmark.RunBenchmark(b, input)
    })
}

func generateTestData(n int) []TestData {
    data := make([]TestData, n)
    for i := 0; i < n; i++ {
        data[i] = TestData{
            ID:   i,
            Data: fmt.Sprintf("test-%d", i),
        }
    }
    return data
}
```

### Latency Benchmark

```go
func BenchmarkNodeLatency(b *testing.B) {
    node := NewTestNode()
    harness := NewBenchmarkHarness(b, node)
    
    b.Run("latency", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            start := time.Now()
            
            harness.SendMessage("input", testMessage)
            _ = harness.ReceiveMessage("output")
            
            latency := time.Since(start)
            harness.metrics.RecordLatency(latency)
        }
    })
}
```

### Resource Usage Benchmark

```go
func BenchmarkResourceUsage(b *testing.B) {
    node := NewTestNode()
    harness := NewBenchmarkHarness(b, node)
    
    b.Run("resources", func(b *testing.B) {
        initialStats := harness.measureResources()
        
        for i := 0; i < b.N; i++ {
            harness.RunWorkload()
            currentStats := harness.measureResources()
            harness.metrics.RecordResourceUsage(currentStats.Delta(initialStats))
        }
    })
}
```

## Performance Profiling

### CPU Profiling

```go
func BenchmarkWithCPUProfile(b *testing.B) {
    f, err := os.Create("cpu.prof")
    if err != nil {
        b.Fatal(err)
    }
    defer f.Close()
    
    if err := pprof.StartCPUProfile(f); err != nil {
        b.Fatal(err)
    }
    defer pprof.StopCPUProfile()
    
    // Run benchmark
    BenchmarkNodeThroughput(b)
}
```

### Memory Profiling

```go
func BenchmarkWithMemoryProfile(b *testing.B) {
    f, err := os.Create("mem.prof")
    if err != nil {
        b.Fatal(err)
    }
    defer f.Close()
    
    // Run benchmark
    BenchmarkNodeThroughput(b)
    
    if err := pprof.WriteHeapProfile(f); err != nil {
        b.Fatal(err)
    }
}
```

## Best Practices

1. **Benchmark Setup**
   - Reset timers appropriately
   - Clear resources between runs
   - Use realistic data sizes
   - Run sufficient iterations

2. **Resource Monitoring**
   - Track memory allocations
   - Monitor goroutine count
   - Measure buffer utilization
   - Watch for leaks

3. **Measurement Accuracy**
   - Account for warmup time
   - Handle outliers
   - Use appropriate sample sizes
   - Consider system load

4. **Performance Goals**
   - Define clear targets
   - Measure relevant metrics
   - Compare against baselines
   - Document limitations

## Common Benchmarking Scenarios

1. **Single Node Performance**
```go
func BenchmarkSingleNode(b *testing.B) {
    node := NewTestNode()
    benchmark := NewBenchmarkNode(node)
    
    b.Run("processing", func(b *testing.B) {
        input := generateTestData(100)
        benchmark.RunBenchmark(b, input)
    })
}
```

2. **Network Flow Performance**
```go
func BenchmarkNetworkFlow(b *testing.B) {
    network := setupTestNetwork()
    benchmark := NewNetworkBenchmark(network)
    
    b.Run("flow", func(b *testing.B) {
        benchmark.RunFlowBenchmark(b, testInput)
    })
}
```

3. **Resource Pool Performance**
```go
func BenchmarkResourcePool(b *testing.B) {
    pool := NewResourcePool(10)
    benchmark := NewPoolBenchmark(pool)
    
    b.Run("pool", func(b *testing.B) {
        benchmark.RunPoolBenchmark(b)
    })
}
```

## Result Analysis

### Metric Collection

```go
type BenchmarkResults struct {
    Throughput      []float64
    Latency        []time.Duration
    ResourceMetrics []ResourceStats
    Errors         []error
}

func (r *BenchmarkResults) Analyze() *BenchmarkAnalysis {
    return &BenchmarkAnalysis{
        AverageThroughput: average(r.Throughput),
        P95Latency:       percentile95(r.Latency),
        MaxMemoryUsage:   maxMemory(r.ResourceMetrics),
        ErrorRate:        float64(len(r.Errors)) / float64(len(r.Throughput)),
    }
}
```

### Performance Reporting

```go
type BenchmarkReport struct {
    NodeID      string
    Timestamp   time.Time
    Analysis    *BenchmarkAnalysis
    Comparison  *BaselineComparison
}

func (r *BenchmarkReport) GenerateReport() string {
    return fmt.Sprintf(`
Performance Report for %s
Time: %v

Throughput: %.2f msg/s
P95 Latency: %v
Max Memory: %v MB
Error Rate: %.2f%%

Baseline Comparison:
- Throughput: %.2f%% of baseline
- Latency: %.2f%% of baseline
- Memory: %.2f%% of baseline
`,
        r.NodeID,
        r.Timestamp,
        r.Analysis.AverageThroughput,
        r.Analysis.P95Latency,
        r.Analysis.MaxMemoryUsage / 1024 / 1024,
        r.Analysis.ErrorRate * 100,
        r.Comparison.ThroughputRatio * 100,
        r.Comparison.LatencyRatio * 100,
        r.Comparison.MemoryRatio * 100,
    )
}
```