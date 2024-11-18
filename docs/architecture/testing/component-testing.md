# Component Testing Architecture

This document details the testing infrastructure for individual components (nodes) in our FBP implementation.

## Core Testing Components

### Node Test Harness

```go
type NodeHarness[In, Out any] struct {
    node        Node
    inputPorts  map[string]chan IP[In]
    outputPorts map[string]chan IP[Out]
    records     *MessageRecorder[Out]
    ctx         context.Context
    cancel      context.CancelFunc
    t           *testing.T
}

func NewNodeHarness[In, Out any](t *testing.T, node Node[In, Out]) *NodeHarness[In, Out] {
    ctx, cancel := context.WithCancel(context.Background())
    return &NodeHarness[In, Out]{
        node:        node,
        inputPorts:  make(map[string]chan IP[In]),
        outputPorts: make(map[string]chan IP[Out]),
        records:     NewMessageRecorder[Out](),
        ctx:         ctx,
        cancel:      cancel,
        t:           t,
    }
}
```

### Message Recorder

```go
type MessageRecorder[T any] struct {
    messages map[string][]T
    mu       sync.RWMutex
    cond     *sync.Cond
}

func NewMessageRecorder[T any]() *MessageRecorder[T] {
    mr := &MessageRecorder[T]{
        messages: make(map[string][]T),
    }
    mr.cond = sync.NewCond(&mr.mu)
    return mr
}
```

### Mock Components

```go
type MockNode[In, Out any] struct {
    mock.Mock
}

func (m *MockNode[In, Out]) Process(ctx context.Context, in Port[In], out Port[Out]) error {
    args := m.Called(ctx, in, out)
    return args.Error(0)
}

type ResourceMock struct {
    mock.Mock
}

func (r *ResourceMock) Initialize(ctx context.Context) error {
    args := r.Called(ctx)
    return args.Error(0)
}
```

## Testing Operations

### Send Test Messages

```go
func (h *NodeHarness[In, Out]) SendMessage(portName string, msg In) error {
    port, ok := h.inputPorts[portName]
    if !ok {
        return fmt.Errorf("port %s not found", portName)
    }

    select {
    case port <- msg:
        return nil
    case <-h.ctx.Done():
        return h.ctx.Err()
    case <-time.After(time.Second):
        return fmt.Errorf("timeout sending message to port %s", portName)
    }
}
```

### Wait for Messages

```go
func (h *NodeHarness[In, Out]) WaitForMessages(portName string, count int) ([]Out, error) {
    return h.records.WaitForMessages(portName, count, time.Second*5)
}

func (mr *MessageRecorder[T]) WaitForMessages(portName string, count int, timeout time.Duration) ([]T, error) {
    done := make(chan struct{})
    var result []T
    
    go func() {
        mr.mu.Lock()
        defer mr.mu.Unlock()
        
        for len(mr.messages[portName]) < count {
            mr.cond.Wait()
        }
        
        result = mr.messages[portName][:count]
        close(done)
    }()
    
    select {
    case <-done:
        return result, nil
    case <-time.After(timeout):
        return nil, fmt.Errorf("timeout waiting for messages on port %s", portName)
    }
}
```

## Testing Patterns

### Basic Component Test

```go
func TestTransformNode(t *testing.T) {
    // Create test harness
    transform := NewTransformer(strings.ToUpper)
    harness := NewNodeHarness(t, transform)
    
    // Test cases
    testCases := []struct{
        input    string
        expected string
    }{
        {"hello", "HELLO"},
        {"world", "WORLD"},
    }
    
    // Run tests
    for _, tc := range testCases {
        assert.NoError(t, harness.SendMessage("in", tc.input))
        results, err := harness.WaitForMessages("out", 1)
        assert.NoError(t, err)
        assert.Equal(t, tc.expected, results[0])
    }
}
```

### Resource-Aware Component Test

```go
func TestDatabaseNode(t *testing.T) {
    // Create network tester
    network := NewNetwork()
    tester := NewNetworkTester(t, network)
    
    // Create mock database
    dbMock := tester.MockResource("db")
    dbMock.On("Initialize", mock.Anything).Return(nil)
    dbMock.On("HealthCheck", mock.Anything).Return(nil)
    dbMock.On("Close", mock.Anything).Return(nil)
    
    // Create mock node
    nodeMock := tester.MockNode[string, string]("dbNode")
    nodeMock.On("Process", mock.Anything, mock.Anything, mock.Anything).Return(nil)
    
    // Run test
    ctx := context.Background()
    assert.NoError(t, network.Run(ctx))
    
    // Verify expectations
    dbMock.AssertExpectations(t)
    nodeMock.AssertExpectations(t)
}
```

## Performance Testing

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

## Best Practices

1. **Test Setup**
   - Create fresh harness for each test
   - Initialize resources properly
   - Clean up after tests
   - Use appropriate timeouts

2. **Message Handling**
   - Test edge cases
   - Verify message ordering
   - Test backpressure scenarios
   - Handle timeouts appropriately

3. **Resource Management**
   - Mock external resources
   - Verify resource cleanup
   - Test resource failures
   - Monitor resource usage

4. **Error Scenarios**
   - Test error propagation
   - Verify error recovery
   - Test cancellation
   - Check cleanup on failures

## Common Testing Scenarios

1. **Basic Processing**
   ```go
   func TestBasicProcessing(t *testing.T) {
       node := NewProcessor()
       harness := NewNodeHarness(t, node)
       
       harness.SendMessage("in", "test")
       results, err := harness.WaitForMessages("out", 1)
       assert.NoError(t, err)
       assert.Equal(t, "TEST", results[0])
   }
   ```

2. **Error Handling**
   ```go
   func TestErrorHandling(t *testing.T) {
       node := NewProcessor()
       harness := NewNodeHarness(t, node)
       
       harness.SendMessage("in", "invalid")
       err := harness.WaitForError()
       assert.Error(t, err)
   }
   ```

3. **Resource Management**
   ```go
   func TestResourceCleanup(t *testing.T) {
       node := NewResourceNode()
       harness := NewNodeHarness(t, node)
       
       ctx := context.Background()
       harness.Run(ctx)
       
       harness.Cancel()
       assert.Eventually(t, func() bool {
           return node.ResourcesClosed()
       }, time.Second, 10*time.Millisecond)
   }
   ```

## Testing Tools

1. **Network Tester**
   - Test full network flows
   - Mock network components
   - Verify network behavior
   - Monitor network metrics

2. **Resource Mocks**
   - Mock external dependencies
   - Simulate failures
   - Control timing
   - Verify interactions

3. **Benchmark Tools**
   - Measure throughput
   - Profile memory usage
   - Identify bottlenecks
   - Compare implementations