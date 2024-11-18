# Initial Information Packets (IIPs) Subsystem

This document details the Initial Information Packets (IIPs) subsystem that provides configuration and initialization data to components in our FBP implementation.

## Core Concepts

### IIP Definition

```go
type IP[T any] struct {
    Type     IPType
    Data     T
    Metadata map[string]any
    Origin   string    // Source tracking
    FlowID   string    // Flow tracking
}

const (
    NormalIP IPType = iota
    InitialIP      // IIP type
    OpenBracket
    CloseBracket
)
```

### Network IIP Management

```go
type Network struct {
    nodes       map[string]interface{}
    connections map[string][]Connection
    iips        map[string]*IP[interface{}]
    mu          sync.RWMutex
}

func (n *Network) AddNode(name string, node interface{}) error {
    n.mu.Lock()
    defer n.mu.Unlock()

    // Validate node
    if err := n.validator.ValidateNode(node); err != nil {
        return fmt.Errorf("invalid node %s: %w", name, err)
    }

    // Add node
    n.nodes[name] = node

    // Process any IIPs defined by the node
    if initializer, ok := node.(interface{ InitialValues() map[string]interface{} }); ok {
        for portName, value := range initializer.InitialValues() {
            iip := &IP[interface{}]{
                Type: InitialIP,
                Data: value,
                Metadata: map[string]interface{}{
                    "targetNode": name,
                    "targetPort": portName,
                },
            }
            n.iips[fmt.Sprintf("%s.%s", name, portName)] = iip
        }
    }

    return nil
}
```

## IIP Delivery System

### IIP Processing

```go
func (n *Network) deliverIIP(targetNode interface{}, targetPort string, iip *IP[interface{}]) error {
    if port, ok := getNodePort(targetNode, targetPort); ok {
        select {
        case port.Channel <- iip:
            return nil
        case <-time.After(time.Second):
            return fmt.Errorf("timeout delivering IIP to port %s", targetPort)
        }
    }
    return fmt.Errorf("port %s not found", targetPort)
}

func (n *Network) Run(ctx context.Context) error {
    // Deliver IIPs before starting normal processing
    for _, iip := range n.iips {
        targetNode := n.nodes[iip.Metadata["targetNode"].(string)]
        targetPort := iip.Metadata["targetPort"].(string)
        if err := n.deliverIIP(targetNode, targetPort, iip); err != nil {
            return fmt.Errorf("delivering IIP: %w", err)
        }
    }

    // Continue with normal network execution...
    return nil
}
```

## Component IIP Support

### IIP-Aware Component

```go
type Component[In, Out any] interface {
    // Initialize is called with IIPs before processing begins
    Initialize(ctx context.Context) error
    
    // Process handles the main processing logic
    Process(ctx context.Context) error
    
    // InitialValues returns IIPs for configuration
    InitialValues() map[string]interface{}
}

// Example component with IIP support
type ConfigurableComponent[In, Out any] struct {
    config    ComponentConfig
    inPort    Port[In]
    outPort   Port[Out]
}

func (c *ConfigurableComponent[In, Out]) InitialValues() map[string]interface{} {
    return map[string]interface{}{
        "config": c.config,
    }
}

func (c *ConfigurableComponent[In, Out]) Initialize(ctx context.Context) error {
    // Process configuration IIP
    select {
    case iip := <-c.inPort.Channel:
        if iip.Type != InitialIP {
            return fmt.Errorf("expected IIP, got %v", iip.Type)
        }
        if cfg, ok := iip.Data.(ComponentConfig); ok {
            c.config = cfg
            return nil
        }
        return fmt.Errorf("invalid config type")
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## IIP Use Cases

### 1. Component Configuration

```go
type DatabaseNode struct {
    ConfigurableComponent[string, string]
}

func NewDatabaseNode() *DatabaseNode {
    return &DatabaseNode{
        ConfigurableComponent: ConfigurableComponent[string, string]{
            config: ComponentConfig{
                "connectionString": "default_connection",
                "poolSize": 10,
                "timeout": "30s",
            },
        },
    }
}

// Network setup with configuration
network.AddNode("db", NewDatabaseNode())
network.SetIIP("db", "config", ComponentConfig{
    "connectionString": "postgresql://...",
    "poolSize": 20,
    "timeout": "1m",
})
```

### 2. Initial State Setting

```go
type StatefulNode struct {
    ConfigurableComponent[string, string]
    state map[string]interface{}
}

func (n *StatefulNode) InitialValues() map[string]interface{} {
    return map[string]interface{}{
        "state": n.state,
        "config": n.config,
    }
}

// Network setup with initial state
network.AddNode("stateful", NewStatefulNode())
network.SetIIP("stateful", "state", map[string]interface{}{
    "counter": 0,
    "lastRun": time.Now(),
})
```

### 3. Connection Parameters

```go
type FilterNode struct {
    ConfigurableComponent[string, string]
    filterCriteria func(string) bool
}

func (n *FilterNode) InitialValues() map[string]interface{} {
    return map[string]interface{}{
        "criteria": n.filterCriteria,
    }
}

// Network setup with filter criteria
network.AddNode("filter", NewFilterNode())
network.SetIIP("filter", "criteria", func(s string) bool {
    return len(s) > 10
})
```

## Best Practices

1. **IIP Design**
   - Keep IIPs simple and serializable
   - Use structured configuration objects
   - Validate IIP data types
   - Handle missing IIPs gracefully

2. **Delivery Timing**
   - Deliver IIPs before normal processing
   - Handle delivery timeouts
   - Verify IIP receipt
   - Clean up after delivery

3. **Configuration Management**
   - Use typed configurations
   - Provide defaults
   - Validate configurations
   - Document required IIPs

4. **Error Handling**
   - Handle missing IIPs
   - Validate IIP types
   - Report configuration errors
   - Maintain consistency

## Common Patterns

### 1. Default Configuration

```go
type ConfiguredNode[In, Out any] struct {
    BaseNode[In, Out]
    config NodeConfig
}

func (n *ConfiguredNode[In, Out]) InitialValues() map[string]interface{} {
    if n.config == nil {
        n.config = DefaultConfig()
    }
    return map[string]interface{}{
        "config": n.config,
    }
}
```

### 2. Configuration Validation

```go
func (n *ConfiguredNode[In, Out]) Initialize(ctx context.Context) error {
    iip := <-n.configPort.Channel
    config, ok := iip.Data.(NodeConfig)
    if !ok {
        return fmt.Errorf("invalid config type")
    }
    
    if err := config.Validate(); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }
    
    n.config = config
    return nil
}
```

### 3. Multiple IIPs

```go
type ComplexNode[In, Out any] struct {
    BaseNode[In, Out]
    config    NodeConfig
    state     NodeState
    resources ResourceConfig
}

func (n *ComplexNode[In, Out]) InitialValues() map[string]interface{} {
    return map[string]interface{}{
        "config": n.config,
        "state": n.state,
        "resources": n.resources,
    }
}
```

## Testing IIPs

```go
func TestIIPDelivery(t *testing.T) {
    node := NewConfigurableNode()
    harness := NewNodeHarness(t, node)
    
    config := NodeConfig{
        "param1": "value1",
        "param2": 42,
    }
    
    // Send IIP
    err := harness.SendIIP("config", config)
    assert.NoError(t, err)
    
    // Verify configuration
    assert.Equal(t, config, node.config)
}
```