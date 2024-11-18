# Port Subsystem

The port subsystem provides type-safe communication channels between processes in our FBP implementation.

## Core Components

### Port Types

```go
type PortType int

const (
    TypeInput PortType = iota
    TypeOutput
)
```

### Port Structure

```go
type Port[T any] struct {
    name        string
    description string
    required    bool
    portType    PortType
    channels    []chan *ip.IP[T]
    maxConns    int
}
```

## Features

### Connection Management
- Default single connection limit
- Configurable connection limits
- Thread-safe connection handling
- Multiple output connections (fan-out)
- Single input connection (fan-in)

### Communication
- Context-aware send/receive operations
- Buffered channels for flow control
- Type-safe data transfer
- Error propagation

### Port Creation

```go
// Input port creation
port := ports.NewInput[string]("name", "description", true)

// Output port creation
port := ports.NewOutput[string]("name", "description", false)
```

### Connection Operations

```go
// Set connection limit
port.SetMaxConnections(2)

// Connect channel
ch := make(chan *ip.IP[string], 1)
err := port.Connect(ch)

// Send data (output ports)
err := port.Send(ctx, packet)

// Receive data (input ports)
packet, err := port.Receive(ctx)
```

## Best Practices

### Port Configuration
- Use meaningful port names
- Provide clear descriptions
- Set appropriate buffer sizes
- Configure connection limits based on needs

### Error Handling
- Check connection errors
- Handle context cancellation
- Verify port connectivity
- Handle timeouts appropriately

### Thread Safety
- Use provided mutex protection
- Don't access channels directly
- Respect port types
- Handle concurrent connections

## Testing

### Test Cases
- Connection limits
- Send/receive operations
- Context cancellation
- Timeout handling
- Error conditions