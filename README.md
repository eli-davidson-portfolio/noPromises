# noPromises: Classical Flow-Based Programming in Go

noPromises is a strict implementation of J. Paul Morrison's Flow-Based Programming (FBP) paradigm in Go. Unlike promise-based or reactive systems, noPromises embraces Go's channel-based concurrency and type system to create truly independent processes that communicate solely through message passing.

The name reflects our philosophy: instead of promises, callbacks, or reactive streams, we use bounded buffers and explicit message passing. Every connection is a contract, every process is independent, and the network is the program.

## Core Philosophy

### True to Classical FBP
noPromises strictly adheres to the fundamental principles of FBP:

1. **Applications as Networks**
   - Programs are networks of black box processes
   - Processes communicate only through pre-defined connections
   - Network structure is separate from process logic
   ```go
   network := nop.NewNetwork().
       AddProcess("reader", NewReader()).
       AddProcess("transform", NewTransformer()).
       AddProcess("writer", NewWriter()).
       Connect("reader", "out", "transform", "in").
       Connect("transform", "out", "writer", "in")
   ```

2. **Processes as "Little Mainlines"**
   - Each process is a true independent unit
   - No shared state or direct process communication
   - Go's goroutines provide real concurrent execution
   ```go
   type Process[In, Out any] interface {
       // Each process runs independently in its own goroutine
       Process(ctx context.Context) error
       InPorts() []Port[In]
       OutPorts() []Port[Out]
   }
   ```

3. **Information Packets (IPs)**
   - All data flows as discrete packets
   - IPs have clear ownership semantics
   - Type-safe implementation using Go generics
   ```go
   type IP[T any] struct {
       Type     IPType         // Normal, Bracket, or Initial
       Data     T             // Type-safe payload
       Metadata map[string]any // Additional information
   }
   ```

4. **Bounded Buffers**
   - Every connection has a fixed capacity
   - Backpressure is explicit and manageable
   - Go channels provide the perfect primitive
   ```go
   type Connection[T any] struct {
       buffer    chan *IP[T]
       capacity  int
       overflow  *list.List  // For backpressure handling
   }
   ```

### Why Go is Perfect for FBP

We leverage Go's unique features to enhance classical FBP:

1. **Goroutines as Natural Processes**
   - Lightweight concurrent execution
   - Built-in scheduling
   - Context-based lifecycle management

2. **Channels as Perfect Connections**
   - Native bounded buffer implementation
   - Built-in synchronization
   - Select-based multi-port handling

3. **Type System as Safety Net**
   - Generic type constraints
   - Compile-time connection validation
   - Interface-based component contracts

4. **Error Model as Process Control**
   - Explicit error handling
   - Graceful failure propagation
   - Resource cleanup guarantees

## Implementation Principles

### 1. Type Safety Above All
```go
// Every process is fully type-safe
type TransformProcess[In, Out any] struct {
    inPort  Port[In]
    outPort Port[Out]
    transform func(In) Out
}
```

### 2. Explicit Port Contracts
```go
type Port[T any] struct {
    Name        string
    Description string
    Required    bool
    Channel     chan *IP[T]
}
```

### 3. Hierarchical Data Support
```go
// Full bracket support for nested structures
type BracketManager[T any] struct {
    depth     int32
    trackers  stack.Stack[BracketType]
}
```

### 4. Resource Safety
```go
// Guaranteed resource cleanup
type ManagedProcess[In, Out any] struct {
    Process[In, Out]
    resources []Resource
    cleanup   func() error
}
```

## Strict Development Rules

### 1. Process Independence
- Zero shared state
- Configuration only through IIPs
- Explicit port contracts
- Resource isolation

### 2. Network Definition
- Compile-time validation
- Explicit connection capacity
- Clear subnet boundaries
- Deadlock prevention

### 3. Testing Requirements
- Process isolation tests
- Network integration tests
- Performance benchmarks
- Race condition verification

### 4. Performance Guidelines
- Explicit buffer sizing
- Resource pooling
- Monitoring hooks
- Backpressure handling

## Current Status

🚧 **Under Active Development** 🚧

Currently focusing on:
1. Core framework implementation
2. Basic process library
3. Network execution engine
4. Testing infrastructure

## Contributing

Contributions must align with our strict FBP principles. See [Contributing Guidelines](CONTRIBUTING.md) before submitting pull requests.

## Design Decisions

### 1. Process Isolation
- No shared memory
- Port-only communication
- No global state
- Clear ownership semantics

### 2. Type System Usage
- Generic constraints
- Interface contracts
- Compile-time validation
- Full type inference

### 3. Resource Control
- Automatic cleanup
- Connection management
- Buffer control
- Lifecycle tracking

### 4. Error Philosophy
- Process-level recovery
- Network error propagation
- Circuit breaking
- Graceful degradation

## License

MIT License - See [LICENSE](LICENSE) for details

## Why noPromises?

1. **True to Classical FBP**
   - Not just dataflow
   - Not reactive streams
   - Real FBP components
   - True process isolation

2. **Go Native**
   - No runtime reflection
   - Natural concurrency
   - Type system power
   - Simple mental model

3. **Production Ready** (Soon)
   - Full test coverage
   - Complete observability
   - Resource safety
   - Performance focused

Stay tuned as we build a pure, principled Flow-Based Programming implementation in Go.
