# Performance Benchmarking

This document outlines benchmarking patterns for our FBP implementation.

## Process Benchmarking

### Basic Process Benchmark
```go
func BenchmarkProcess(b *testing.B) {
    proc := NewTestProcess()
    ctx := context.Background()
    
    inCh := make(chan *ip.IP[string], 1)
    outCh := make(chan *ip.IP[string], 1)
    
    require.NoError(b, proc.in.Connect(inCh))
    require.NoError(b, proc.out.Connect(outCh))
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        inCh <- ip.New[string]("test")
        <-outCh
    }
}
```

## Network Benchmarking

### Network Flow Benchmark
```go
func BenchmarkNetworkFlow(b *testing.B) {
    net := network.New()
    proc1 := NewTestProcess()
    proc2 := NewTestProcess()
    
    net.AddProcess("proc1", proc1)
    net.AddProcess("proc2", proc2)
    net.Connect("proc1", "out", "proc2", "in")
    
    inputCh := make(chan *ip.IP[string], b.N)
    outputCh := make(chan *ip.IP[string], b.N)
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        inputCh <- ip.New[string]("test")
        <-outputCh
    }
}
```

## Best Practices

### Benchmark Setup
- Use buffered channels
- Reset timer after setup
- Clean up resources
- Handle context properly

### Measurement Points
- Process throughput
- Network latency
- Resource usage
- Error rates