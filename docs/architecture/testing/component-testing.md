# Component Testing Guide

This document outlines testing patterns for our FBP implementation components.

## Process Testing

### Basic Process Test Structure

```go
type TestProcess struct {
    process.BaseProcess
    in  *ports.Port[string]
    out *ports.Port[string]
}

func TestProcessLifecycle(t *testing.T) {
    t.Run("initialization", func(t *testing.T) {
        proc := NewTestProcess()
        ctx := context.Background()
        
        err := proc.Initialize(ctx)
        require.NoError(t, err)
        assert.True(t, proc.IsInitialized())
    })
}
```

### Testing Data Flow

```go
func TestProcessing(t *testing.T) {
    t.Run("basic transformation", func(t *testing.T) {
        proc := NewTestProcess()
        ctx := context.Background()
        
        inCh := make(chan *ip.IP[string], 1)
        outCh := make(chan *ip.IP[string], 1)
        
        require.NoError(t, proc.in.Connect(inCh))
        require.NoError(t, proc.out.Connect(outCh))
        
        go func() {
            err := proc.Process(ctx)
            require.NoError(t, err)
        }()
        
        inCh <- ip.New("test")
        result := <-outCh
        assert.Equal(t, "TEST", result.Data())
    })
}
```

## Network Testing

### Network Creation Tests

```go
func TestNetwork(t *testing.T) {
    t.Run("creation", func(t *testing.T) {
        net := network.New()
        assert.NotNil(t, net)
        assert.Equal(t, 0, net.ProcessCount())
    })
}
```

### Connection Tests

```go
t.Run("valid connection", func(t *testing.T) {
    net := network.New()
    proc1 := NewTestProcess()
    proc2 := NewTestProcess()
    
    require.NoError(t, net.AddProcess("proc1", proc1))
    require.NoError(t, net.AddProcess("proc2", proc2))
    
    err := net.Connect("proc1", "out", "proc2", "in")
    assert.NoError(t, err)
})
```

### Execution Tests

```go
t.Run("successful execution", func(t *testing.T) {
    net := network.New()
    proc1 := NewTestProcess()
    proc2 := NewTestProcess()
    
    // Setup network
    require.NoError(t, net.AddProcess("proc1", proc1))
    require.NoError(t, net.AddProcess("proc2", proc2))
    require.NoError(t, net.Connect("proc1", "out", "proc2", "in"))
    
    // Create test channels
    inputCh := make(chan *ip.IP[string], 1)
    outputCh := make(chan *ip.IP[string], 1)
    
    // Run network
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    networkDone := make(chan error, 1)
    go func() {
        networkDone <- net.Run(ctx)
    }()
    
    // Test data flow
    testData := "test"
    inputCh <- ip.New[string](testData)
    
    select {
    case result := <-outputCh:
        assert.Equal(t, strings.ToUpper(testData), result.Data())
    case <-time.After(time.Second):
        t.Fatal("Test timed out")
    }
})
```

## Best Practices

### Test Setup
- Use context for cancellation
- Create buffered channels
- Clean up resources
- Handle timeouts

### Error Testing
- Test error propagation
- Test cleanup on failure
- Test context cancellation
- Test invalid configurations

### Concurrency Testing
- Use race detector
- Test concurrent operations
- Handle timeouts properly
- Test cleanup in all cases