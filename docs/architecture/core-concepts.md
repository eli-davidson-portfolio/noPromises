# noPromises Architecture

## Overview

noPromises is a strict implementation of Flow-Based Programming (FBP) in Go, emphasizing compile-time correctness, true concurrency, and adherence to classical FBP principles. This document outlines the architectural decisions, core components, and implementation guidelines.

## Core Principles

### 1. Classical FBP Adherence
- **Processes as Black Boxes**: Each process is fully independent
- **Port-Based Communication**: All data flow through typed ports
- **Information Packets**: Data moves as discrete, owned packets
- **Bounded Buffers**: All connections have fixed capacity
- **External Coordination**: Network structure defined separately from processes

### 2. Go Native Implementation
- **Goroutines as Processes**: Each FBP process is a goroutine
- **Channels as Connections**: Native bounded buffer implementation
- **Type System**: Generic constraints and interface contracts
- **Error Model**: Explicit error handling and propagation
- **Context Usage**: Lifecycle and cancellation management

## Core Components

### 1. Information Packets (IPs)
```go
type IP[T any] struct {
    Type     IPType         // Normal, Bracket, Initial
    Data     T             // Type-safe payload
    Metadata map[string]any // Additional information
}
```

- **Ownership Semantics**: Only one process owns an IP at a time
- **Type Safety**: Generic implementation ensures type correctness
- **Metadata Support**: Extensible metadata for debugging and tracking
- **Special Types**: Support for brackets and Initial Information Packets (IIPs)

### 2. Ports
```go
type Port[T any] interface {
    Send(ctx context.Context, data T) error
    Receive(ctx context.Context) (T, error)
    Close() error
    Name() string
    IsConnected() bool
}
```

- **Direction Specific**: Input and output ports are distinct
- **Type Safe**: Generic type parameters ensure compatibility
- **Context Aware**: Support for cancellation and deadlines
- **Connection Status**: Explicit connection state management

### 3. Processes
```go
type Process[In, Out any] interface {
    Initialize(ctx context.Context) error
    Process(ctx context.Context) error
    Shutdown(ctx context.Context) error
    InPorts() []Port[In]
    OutPorts() []Port[Out]
}
```

- **Lifecycle Management**: Clear initialization and shutdown phases
- **Port Declaration**: Explicit port definitions
- **Resource Safety**: Guaranteed cleanup through shutdown phase
- **Error Handling**: Comprehensive error management

### 4. Networks
```go
type Network interface {
    AddProcess(name string, process Process) error
    Connect(fromProcess, fromPort, toProcess, toPort string) error
    Run(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

- **Process Management**: Addition and removal of processes
- **Connection Management**: Port connection and disconnection
- **Validation**: Compile-time network validation where possible
- **Runtime Control**: Start, stop, and monitoring capabilities

## Implementation Details

### 1. Resource Management
- **Connection Pooling**: Reuse of connection resources
- **Buffer Management**: Dynamic buffer sizing based on pressure
- **Resource Cleanup**: Deterministic resource cleanup
- **Memory Safety**: No resource leaks through careful tracking

### 2. Error Handling
- **Process Level**: Individual process error management
- **Network Level**: Error propagation through network
- **Circuit Breaking**: Automatic failure handling
- **Recovery Strategies**: Configurable recovery policies

### 3. Concurrency Management
- **Goroutine Control**: Careful goroutine lifecycle management
- **Channel Usage**: Proper channel closure and cleanup
- **Context Usage**: Cancellation and timeout management
- **Deadlock Prevention**: Active detection and prevention

### 4. Performance Considerations
- **Buffer Sizing**: Automatic or configurable buffer capacities
- **Backpressure**: Explicit handling of buffer pressure
- **Resource Pooling**: Connection and resource reuse
- **Monitoring**: Built-in performance metrics

## Design Patterns

### 1. Process Patterns
- **Transform Process**: Basic data transformation
- **Filter Process**: Data filtering and routing
- **Merge Process**: Combining multiple inputs
- **Split Process**: Distributing to multiple outputs

### 2. Network Patterns
- **Pipeline**: Linear process chains
- **Fan-Out**: Distribution patterns
- **Fan-In**: Aggregation patterns
- **Feedback Loops**: Controlled cyclic flows

### 3. Error Patterns
- **Retry Pattern**: Automatic retry with backoff
- **Circuit Breaker**: Failure protection
- **Dead Letter Channel**: Error message handling
- **Compensating Actions**: Error recovery

## Testing Strategy

### 1. Unit Testing
- **Process Testing**: Individual process behavior
- **Port Testing**: Port contract verification
- **IP Testing**: Packet handling verification
- **Error Testing**: Error condition handling

### 2. Integration Testing
- **Network Testing**: Full network behavior
- **Flow Testing**: Data flow verification
- **Error Propagation**: Error handling paths
- **Resource Management**: Resource cleanup verification

### 3. Performance Testing
- **Benchmark Suite**: Standard performance tests
- **Load Testing**: High-volume data handling
- **Resource Usage**: Memory and goroutine tracking
- **Deadlock Detection**: Concurrent behavior verification

## Security Considerations

### 1. Resource Protection
- **Bounded Memory**: Fixed buffer sizes
- **Cleanup Guarantees**: Resource cleanup enforcement
- **Isolation**: Process separation
- **Input Validation**: Port-level validation

### 2. Concurrency Safety
- **Race Detection**: Built-in race condition checking
- **Deadlock Prevention**: Active deadlock detection
- **Resource Limits**: Explicit resource boundaries
- **Cancellation**: Proper shutdown procedures

## Future Considerations

### 1. Extensions
- **Distribution**: Network distribution support
- **Persistence**: State persistence options
- **Monitoring**: Enhanced monitoring capabilities
- **Visualization**: Network visualization tools

### 2. Optimizations
- **Buffer Tuning**: Automatic buffer optimization
- **Resource Pooling**: Enhanced resource reuse
- **Scheduling**: Advanced scheduling strategies
- **Memory Management**: Reduced allocation patterns

## Versioning and Compatibility

### 1. Version Strategy
- **Semantic Versioning**: Clear version progression
- **API Stability**: Interface stability guarantees
- **Migration Support**: Version migration tools
- **Deprecation Policy**: Clear deprecation process

### 2. Compatibility Guarantees
- **API Compatibility**: Clear API version guarantees
- **Network Compatibility**: Network definition stability
- **Process Compatibility**: Process contract stability
- **Upgrade Path**: Clear upgrade procedures