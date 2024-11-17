# GoFlows: Classical Flow-Based Programming for Go

GoFlows is a principled implementation of J. Paul Morrison's Flow-Based Programming (FBP) paradigm in Go, emphasizing type safety, true concurrency, and compile-time guarantees. We're not just building another dataflow framework; we're creating a system that genuinely embodies the core principles of FBP while leveraging Go's natural strengths in concurrent programming.

## Core Philosophy

### True to Classical FBP
GoFlows strictly adheres to the fundamental principles of FBP:

1. **Applications as Networks**
   - Programs are designed as networks of black box processes
   - Processes exchange data through pre-defined connections
   - Networks are defined separately from processes
   ```go
   network := flows.NewNetwork().
       AddProcess("reader", NewReader()).
       AddProcess("transform", NewTransformer()).
       AddProcess("writer", NewWriter()).
       Connect("reader", "out", "transform", "in").
       Connect("transform", "out", "writer", "in")
   ```

2. **Processes as "Little Mainlines"**
   - Each process runs independently
   - Processes are unaware of each other
   - Go's goroutines provide true concurrent execution
   ```go
   type Process[In, Out any] interface {
       // Process runs independently in its own goroutine
       Process(ctx context.Context) error
       InPorts() []Port[In]
       OutPorts() []Port[Out]
   }
   ```

3. **Information Packets (IPs)**
   - Data flows as discrete packets
   - IPs are owned by one process at a time
   - Type-safe implementation using Go generics
   ```go
   type IP[T any] struct {
       Type     IPType         // Normal, Bracket, or Initial
       Data     T             // Type-safe payload
       Metadata map[string]any // Additional information
   }
   ```

4. **Bounded Buffers**
   - All connections have fixed capacity
   - Backpressure is handled explicitly
   - Go channels provide natural implementation
   ```go
   type Connection[T any] struct {
       buffer    chan *IP[T]
       capacity  int
       overflow  *list.List  // For backpressure handling
   }
   ```

### Go's Natural Fit

We leverage Go's strengths to enhance classical FBP:

1. **Goroutines as Processes**
   - Each FBP process runs in its own goroutine
   - Scheduler handles true parallel execution
   - Context-based lifecycle management

2. **Channels as Connections**
   - Go channels implement bounded buffers naturally
   - Select statements for multi-port handling
   - Built-in synchronization primitives

3. **Type Safety with Generics**
   - Compile-time type checking for IPs
   - Type-safe ports and connections
   - Generic process implementations

4. **Error Handling**
   - Go's error model for process failures
   - Graceful error propagation
   - Circuit breakers and recovery patterns

## Implementation Principles

### 1. Strict Type Safety
```go
// Example of type-safe process definition
type TransformProcess[In, Out any] struct {
    inPort  Port[In]
    outPort Port[Out]
    transform func(In) Out
}
```

### 2. Explicit Port Definitions
```go
type Port[T any] struct {
    Name        string
    Description string
    Required    bool
    Channel     chan *IP[T]
}
```

### 3. Bracket Support
```go
// Support for hierarchical data structures
type BracketManager[T any] struct {
    depth     int32
    trackers  stack.Stack[BracketType]
}
```

### 4. Resource Management
```go
// Automatic resource cleanup
type ManagedProcess[In, Out any] struct {
    Process[In, Out]
    resources []Resource
    cleanup   func() error
}
```

## Development Guidelines

### 1. Process Development
- Processes must be completely independent
- All configuration through Initial Information Packets (IIPs)
- No shared state between processes
- Explicit error handling and resource management

### 2. Network Definition
- Networks defined separately from process logic
- Compile-time validation of connections
- Support for hierarchical networks (substreams)
- Dynamic network modification where necessary

### 3. Testing Requirements
- Every process must have isolated tests
- Network-level integration tests
- Performance benchmarks for critical paths
- Deadlock and race condition testing

### 4. Performance Considerations
- Buffering strategies must be explicit
- Resource pools for expensive operations
- Monitoring and metrics built-in
- Deadlock prevention mechanisms

## Current Status

🚧 **Under Active Development** 🚧

We are currently in the early stages of development, focusing on:
1. Core framework implementation
2. Basic process types
3. Network execution engine
4. Testing infrastructure

## Contributing

We welcome contributions that align with our core philosophy. Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting pull requests.

## Design Decisions

We maintain strict adherence to FBP principles through:

1. **Isolation**
   - Processes cannot share memory
   - All communication through ports
   - No global state

2. **Type Safety**
   - Compile-time type checking
   - Generic process implementations
   - Type-safe network definitions

3. **Resource Management**
   - Automatic cleanup
   - Connection pooling
   - Bounded buffer enforcement

4. **Error Handling**
   - Process-level recovery
   - Network-level error propagation
   - Circuit breaker patterns

## License

MIT License - See [LICENSE](LICENSE) for details

## Why GoFlows?

1. **True to FBP**
   - Not just a dataflow framework
   - Complete implementation of FBP concepts
   - Support for all FBP patterns

2. **Go Native**
   - Leverages Go's concurrency model
   - Type-safe implementations
   - Excellent performance characteristics

3. **Production Ready** (Soon)
   - Comprehensive testing
   - Full observability
   - Resource management
   - Error handling

Stay tuned for updates as we build a principled, production-grade FBP implementation in Go.
