# Core Concepts: Flow-Based Programming in Go

This document outlines the fundamental architectural concepts of our Flow-Based Programming (FBP) implementation in Go.

## Core Components

### Information Packets (IPs)

IPs are the fundamental unit of data flowing through the network:

```go
type IP[T any] struct {
    Data     T
    Metadata map[string]any
}
```

Key features:
- Generic type support
- Metadata storage
- Thread-safe operations
- Owner tracking

### Ports

Ports are the connection points between processes:

```go
type Port[T any] struct {
    name        string
    description string
    required    bool
    portType    PortType
    channels    []chan *IP[T]
    maxConns    int
}
```

Features:
- Type-safe connections
- Connection limits
- Buffered channels
- Context-aware operations
- Support for fan-out (output ports)
- Support for fan-in (input ports)

### Processes

Processes are the computational units:

```go
type Process interface {
    Initialize(ctx context.Context) error
    Process(ctx context.Context) error
    Shutdown(ctx context.Context) error
    IsInitialized() bool
}
```

Features:
- Context-aware lifecycle
- Clean initialization/shutdown
- State management
- Error propagation
- Port management

### Network

Networks orchestrate process execution:

```go
type Network struct {
    processes map[string]Process
}
```

Features:
- Process management
- Connection management
- Error handling
- Context-based control
- Clean shutdown

## Key Design Patterns

### Process Lifecycle
1. Initialization
2. Processing
3. Shutdown

### Error Handling
- Context cancellation
- Process errors
- Connection errors
- Cleanup on failure

### Connection Management
- Type-safe channels
- Buffered connections
- Connection limits
- Fan-out support

### Concurrency Model
- Process isolation
- Channel-based communication
- Context-based cancellation
- Synchronized state access

## Best Practices

### Process Implementation
```go
type CustomProcess struct {
    process.BaseProcess
    in  *ports.Port[InputType]
    out *ports.Port[OutputType]
}

func (p *CustomProcess) Process(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Process data
        }
    }
}
```

### Network Configuration
```go
net := network.New()
net.AddProcess("proc1", NewProcess1())
net.AddProcess("proc2", NewProcess2())
net.Connect("proc1", "out", "proc2", "in")
```

### Error Handling
```go
if err := net.Run(ctx); err != nil {
    // Handle network error
    // All processes will be properly shut down
}
```