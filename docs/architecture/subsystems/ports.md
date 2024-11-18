# Ports Subsystem Architecture

This document details the ports subsystem that enables type-safe communication between components in our FBP implementation.

## Core Components

### Port Definition

```go
type Port[T any] struct {
    Name        string
    Description string
    Required    bool
    Channel     chan *IP[T]
    MaxInputs   int  // For many-to-one connections
    MaxOutputs  int  // For one-to-many connections
    connected   bool
    mu          sync.RWMutex
}

func NewPort[T any](name string, required bool) *Port[T] {
    return &Port[T]{
        Name:        name,
        Required:    required,
        Channel:     make(chan *IP[T], 1000), // Default buffer size
        MaxInputs:   1,  // Default to one-to-one
        MaxOutputs:  1,
        Metadata:    make(map[string]interface{}),
    }
}
```

### Port Management

```go
type Node[In, Out any] interface {
    // Port management
    InPorts() []Port[In]
    OutPorts() []Port[Out]
    
    // Port connection validation
    ValidateConnections() error
}

// Base implementation of port management
type BaseNode[In, Out any] struct {
    inPorts     []*Port[In]
    outPorts    []*Port[Out]
    mu          sync.RWMutex
}

func (n *BaseNode[In, Out]) InPorts() []Port[In] {
    n.mu.RLock()
    defer n.mu.RUnlock()
    return n.inPorts
}

func (n *BaseNode[In, Out]) OutPorts() []Port[Out] {
    n.mu.RLock()
    defer n.mu.RUnlock()
    return n.outPorts
}
```

## Port Operations

### Port Connection

```go
func (p *Port[T]) Connect(other *Port[T]) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    // Validate connection
    if p.connected && p.MaxInputs <= 1 {
        return fmt.Errorf("port %s already connected", p.Name)
    }

    if len(p.Channel) > 0 {
        return fmt.Errorf("cannot connect port %s with pending messages", p.Name)
    }

    p.connected = true
    return nil
}

func (p *Port[T]) Disconnect() error {
    p.mu.Lock()
    defer p.mu.Unlock()

    if !p.connected {
        return fmt.Errorf("port %s not connected", p.Name)
    }

    close(p.Channel)
    p.connected = false
    return nil
}
```

### Message Handling

```go
func (p *Port[T]) Send(ctx context.Context, msg *IP[T]) error {
    if !p.connected {
        return fmt.Errorf("port %s not connected", p.Name)
    }

    select {
    case p.Channel <- msg:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Handle backpressure
        return fmt.Errorf("port %s buffer full", p.Name)
    }
}

func (p *Port[T]) Receive(ctx context.Context) (*IP[T], error) {
    if !p.connected {
        return nil, fmt.Errorf("port %s not connected", p.Name)
    }

    select {
    case msg, ok := <-p.Channel:
        if !ok {
            return nil, fmt.Errorf("port %s closed", p.Name)
        }
        return msg, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

## Connection Types

### One-to-One Connection

```go
type OneToOneConnection[T any] struct {
    source      *Port[T]
    destination *Port[T]
    buffer      chan *IP[T]
}

func NewOneToOneConnection[T any](source, dest *Port[T], bufferSize int) *OneToOneConnection[T] {
    return &OneToOneConnection[T]{
        source:      source,
        destination: dest,
        buffer:      make(chan *IP[T], bufferSize),
    }
}
```

### Many-to-One Connection

```go
type ManyToOneConnection[T any] struct {
    sources     []*Port[T]
    destination *Port[T]
    buffer      chan *IP[T]
    mu          sync.RWMutex
}

func (c *ManyToOneConnection[T]) AddSource(source *Port[T]) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if len(c.sources) >= c.destination.MaxInputs {
        return fmt.Errorf("maximum inputs exceeded for port %s", c.destination.Name)
    }

    c.sources = append(c.sources, source)
    return nil
}
```

### One-to-Many Connection

```go
type OneToManyConnection[T any] struct {
    source        *Port[T]
    destinations  []*Port[T]
    buffer        chan *IP[T]
    mu            sync.RWMutex
}

func (c *OneToManyConnection[T]) AddDestination(dest *Port[T]) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if len(c.destinations) >= c.source.MaxOutputs {
        return fmt.Errorf("maximum outputs exceeded for port %s", c.source.Name)
    }

    c.destinations = append(c.destinations, dest)
    return nil
}
```

## Port Validation

### Connection Validation

```go
func (n *BaseNode[In, Out]) ValidateConnections() error {
    // Validate required ports are connected
    for _, port := range n.inPorts {
        if port.Required && !port.connected {
            return fmt.Errorf("required input port %s not connected", port.Name)
        }
    }

    for _, port := range n.outPorts {
        if port.Required && !port.connected {
            return fmt.Errorf("required output port %s not connected", port.Name)
        }
    }

    return nil
}
```

### Type Validation

```go
func ValidatePortTypes[T any](source, dest *Port[T]) error {
    sourceType := reflect.TypeOf((*T)(nil)).Elem()
    destType := reflect.TypeOf((*T)(nil)).Elem()

    if sourceType != destType {
        return fmt.Errorf(
            "type mismatch: source port %s (%v) cannot connect to destination port %s (%v)",
            source.Name, sourceType, dest.Name, destType,
        )
    }

    return nil
}
```

## Port Patterns

### Array Port

```go
type ArrayPort[T any] struct {
    Port[T]
    connections []*Connection[T]
    index      int
}

func (p *ArrayPort[T]) NextConnection() *Connection[T] {
    conn := p.connections[p.index]
    p.index = (p.index + 1) % len(p.connections)
    return conn
}
```

### Optional Port

```go
type OptionalPort[T any] struct {
    Port[T]
    defaultValue T
}

func (p *OptionalPort[T]) Receive(ctx context.Context) (*IP[T], error) {
    if !p.connected {
        return &IP[T]{Data: p.defaultValue}, nil
    }
    return p.Port.Receive(ctx)
}
```

## Best Practices

1. **Port Design**
   - Use descriptive port names
   - Document port purposes
   - Set appropriate buffer sizes
   - Mark required ports

2. **Connection Management**
   - Validate connections early
   - Handle disconnections gracefully
   - Implement proper cleanup
   - Monitor buffer usage

3. **Type Safety**
   - Use generics consistently
   - Validate type compatibility
   - Handle type conversions
   - Document type requirements

4. **Error Handling**
   - Handle connection errors
   - Manage backpressure
   - Clean up on failures
   - Report port status

## Common Use Cases

### 1. Basic Data Flow

```go
type ProcessNode[In, Out any] struct {
    BaseNode[In, Out]
    processor func(In) Out
}

func NewProcessNode[In, Out any](processor func(In) Out) *ProcessNode[In, Out] {
    return &ProcessNode[In, Out]{
        BaseNode: BaseNode[In, Out]{
            inPorts: []*Port[In]{NewPort[In]("in", true)},
            outPorts: []*Port[Out]{NewPort[Out]("out", true)},
        },
        processor: processor,
    }
}
```

### 2. Multiple Inputs

```go
type MergeNode[T any] struct {
    BaseNode[T, T]
}

func NewMergeNode[T any]() *MergeNode[T] {
    return &MergeNode[T]{
        BaseNode: BaseNode[T, T]{
            inPorts: []*Port[T]{
                NewPort[T]("in1", true),
                NewPort[T]("in2", true),
            },
            outPorts: []*Port[T]{NewPort[T]("out", true)},
        },
    }
}
```

### 3. Multiple Outputs

```go
type SplitNode[T any] struct {
    BaseNode[T, T]
}

func NewSplitNode[T any]() *SplitNode[T] {
    return &SplitNode[T]{
        BaseNode: BaseNode[T, T]{
            inPorts: []*Port[T]{NewPort[T]("in", true)},
            outPorts: []*Port[T]{
                NewPort[T]("out1", true),
                NewPort[T]("out2", true),
            },
        },
    }
}
```

## Testing Ports

```go
func TestPortConnection(t *testing.T) {
    source := NewPort[string]("source", true)
    dest := NewPort[string]("dest", true)

    err := source.Connect(dest)
    assert.NoError(t, err)
    assert.True(t, source.connected)
    assert.True(t, dest.connected)

    msg := &IP[string]{Data: "test"}
    ctx := context.Background()

    err = source.Send(ctx, msg)
    assert.NoError(t, err)

    received, err := dest.Receive(ctx)
    assert.NoError(t, err)
    assert.Equal(t, msg, received)
}
```