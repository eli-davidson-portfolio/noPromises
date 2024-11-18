# Getting Started with noPromises

Welcome to noPromises, a principled Flow-Based Programming (FBP) framework for Go. This guide will walk you through the basic concepts and help you create your first FBP application.

## Installation

```bash
go get github.com/noPromises/noPromises
```

## Core Concepts

Before diving in, let's understand the key concepts of Flow-Based Programming:

1. **Processes**: Independent components that process data
2. **Ports**: Connection points where data flows in or out of processes
3. **Connections**: Links between ports that carry data
4. **Information Packets (IPs)**: Discrete units of data flowing through the network

## Your First Flow

Let's create a simple flow that reads a file, transforms its content to uppercase, and writes it to another file.

```go
package main

import (
    "context"
    "github.com/noPromises/noPromises"
)

func main() {
    // Create a new network
    network := nop.NewNetwork()

    // Add processes to the network
    network.AddProcess("reader", NewFileReader("input.txt"))
    network.AddProcess("transform", NewTransformer(strings.ToUpper))
    network.AddProcess("writer", NewFileWriter("output.txt"))

    // Connect the processes
    network.Connect("reader", "out", "transform", "in")
    network.Connect("transform", "out", "writer", "in")

    // Run the network
    if err := network.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

## Creating Custom Processes

Here's how to create your own process:

```go
// Custom process that transforms input data
type TransformProcess[In, Out any] struct {
    inPort     *nop.Port[In]
    outPort    *nop.Port[Out]
    transform  func(In) Out
}

func NewTransformer[In, Out any](fn func(In) Out) *TransformProcess[In, Out] {
    return &TransformProcess[In, Out]{
        inPort:    nop.NewPort[In]("in", true),
        outPort:   nop.NewPort[Out]("out", true),
        transform: fn,
    }
}

func (p *TransformProcess[In, Out]) Process(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case ip := <-p.inPort.Channel:
            result := p.transform(ip.Data)
            p.outPort.Channel <- &nop.IP[Out]{
                Type: nop.NormalIP,
                Data: result,
            }
        }
    }
}

func (p *TransformProcess[In, Out]) InPorts() []*nop.Port[In] {
    return []*nop.Port[In]{p.inPort}
}

func (p *TransformProcess[In, Out]) OutPorts() []*nop.Port[Out] {
    return []*nop.Port[Out]{p.outPort}
}
```

## Error Handling

noPromises provides comprehensive error handling:

```go
// Add error handling to your process
type SafeProcess[In, Out any] struct {
    *TransformProcess[In, Out]
    errorHandler *nop.ErrorHandler
}

func NewSafeProcess[In, Out any](process *TransformProcess[In, Out]) *SafeProcess[In, Out] {
    return &SafeProcess[In, Out]{
        TransformProcess: process,
        errorHandler:    nop.NewErrorHandler(),
    }
}

func (p *SafeProcess[In, Out]) Process(ctx context.Context) error {
    // Add error recovery
    defer func() {
        if r := recover(); r != nil {
            p.errorHandler.HandleError(&nop.NodeError{
                Severity: nop.SeverityFatal,
                Err:      fmt.Errorf("panic: %v", r),
            })
        }
    }()

    return p.TransformProcess.Process(ctx)
}
```

## Testing Your Flows

noPromises includes a comprehensive testing framework:

```go
func TestTransformProcess(t *testing.T) {
    // Create test harness
    transform := NewTransformer(strings.ToUpper)
    harness := nop.NewNodeHarness(t, transform)

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
        harness.SendMessage("in", tc.input)
        result, err := harness.WaitForMessages("out", 1)
        assert.NoError(t, err)
        assert.Equal(t, tc.expected, result[0])
    }
}
```

## Monitoring and Debugging

Add monitoring to your processes:

```go
// Create monitored version of your process
collector := nop.NewMetricsCollector("my-app")
monitored, err := nop.NewMonitoredNode("transform", transform, collector)
if err != nil {
    log.Fatal(err)
}

// Access metrics
metrics := collector.GetNodeMetrics("transform")
fmt.Printf("Messages processed: %d\n", metrics.MessagesProcessed)
fmt.Printf("Average processing time: %v\n", metrics.AverageProcessingTime)
```

## Best Practices

1. **Process Design**
   - Keep processes focused on a single responsibility
   - Use clear, descriptive port names
   - Handle context cancellation properly
   - Implement proper resource cleanup

2. **Network Design**
   - Use appropriate buffer sizes for connections
   - Monitor for potential deadlocks
   - Implement error recovery strategies
   - Document network topology

3. **Testing**
   - Test processes in isolation
   - Test full network flows
   - Include performance benchmarks
   - Test error conditions

4. **Resource Management**
   - Always clean up resources
   - Use connection pools for expensive resources
   - Monitor buffer utilization
   - Implement backpressure handling

## Common Patterns

### 1. Pipeline Pattern
```go
network.AddProcess("step1", NewStep1())
network.AddProcess("step2", NewStep2())
network.AddProcess("step3", NewStep3())

network.Connect("step1", "out", "step2", "in")
network.Connect("step2", "out", "step3", "in")
```

### 2. Fan-Out Pattern
```go
network.AddProcess("source", NewSource())
network.AddProcess("worker1", NewWorker())
network.AddProcess("worker2", NewWorker())

network.Connect("source", "out", "worker1", "in")
network.Connect("source", "out", "worker2", "in")
```

### 3. Aggregation Pattern
```go
network.AddProcess("source1", NewSource())
network.AddProcess("source2", NewSource())
network.AddProcess("merger", NewMerger())

network.Connect("source1", "out", "merger", "in1")
network.Connect("source2", "out", "merger", "in2")
```

## Next Steps

1. Explore the [examples directory](examples/) for more complex flows
2. Read the [API documentation](docs/api.md) for detailed reference
3. Join our [community](CONTRIBUTING.md) to contribute
4. Check out the [performance tuning guide](docs/performance.md)

## Need Help?

- File an issue on GitHub
- Join our Discord community
- Check our FAQ section
- Read our detailed documentation

Remember: noPromises is designed to be explicit and type-safe. If something feels wrong or complicated, there's probably a simpler way to do it. Don't hesitate to ask for help!