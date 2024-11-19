# Server Engine Interface

## Overview

The Server Engine provides a clean interface between the HTTP server and the core FBP system. This interface ensures complete decoupling between the network management logic and the HTTP API layer.

## Core Interface

```go
// FlowEngine defines the interface for managing FBP networks
type FlowEngine interface {
    // Flow Management
    CreateFlow(config FlowConfig) (string, error)
    StartFlow(id string) error
    StopFlow(id string) error
    DeleteFlow(id string) error
    GetFlowStatus(id string) (FlowStatus, error)
    ListFlows() ([]FlowSummary, error)
    
    // Process Discovery
    ListAvailableProcessTypes() ([]ProcessType, error)
}
```

## Type Definitions

### Flow Configuration

```go
// FlowConfig defines the structure for creating a new flow
type FlowConfig struct {
    ID       string                `json:"id"`
    Nodes    map[string]NodeConfig `json:"nodes"`
    Edges    []EdgeConfig         `json:"edges"`
    Metadata map[string]any       `json:"metadata,omitempty"`
}

// NodeConfig defines a single node in the flow
type NodeConfig struct {
    Type   string         `json:"type"`
    Config map[string]any `json:"config"`
}

// EdgeConfig defines a connection between nodes
type EdgeConfig struct {
    From NodePort `json:"from"`
    To   NodePort `json:"to"`
}

// NodePort identifies a specific port on a node
type NodePort struct {
    Node string `json:"node"`
    Port string `json:"port"`
}
```

### Status Information

```go
// FlowStatus provides the current state of a flow
type FlowStatus struct {
    ID          string    `json:"id"`
    State       string    `json:"state"`  // "created", "starting", "running", "stopping", "stopped", "error"
    StartedAt   time.Time `json:"started_at,omitempty"`
    Error       string    `json:"error,omitempty"`
    NodeStatus  map[string]NodeStatus `json:"nodes"`
}

// NodeStatus provides status information for a single node
type NodeStatus struct {
    State             string    `json:"state"`
    MessagesProcessed int64     `json:"messages_processed"`
    LastActive        time.Time `json:"last_active"`
    Error            string    `json:"error,omitempty"`
}

// FlowSummary provides brief information about a flow
type FlowSummary struct {
    ID      string    `json:"id"`
    State   string    `json:"state"`
    Nodes   int       `json:"node_count"`
    Created time.Time `json:"created_at"`
}
```

### Process Information

```go
// ProcessType describes an available process type
type ProcessType struct {
    Name        string       `json:"name"`
    Description string       `json:"description"`
    InputPorts  []PortInfo   `json:"input_ports"`
    OutputPorts []PortInfo   `json:"output_ports"`
    Config      ConfigSchema `json:"config"`
}

// PortInfo describes a port on a process
type PortInfo struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Required    bool   `json:"required"`
}

// ConfigSchema describes the configuration options for a process
type ConfigSchema struct {
    Properties map[string]PropertySchema `json:"properties"`
    Required   []string                 `json:"required"`
}

// PropertySchema describes a single configuration property
type PropertySchema struct {
    Type        string `json:"type"`
    Description string `json:"description"`
    Default     any    `json:"default,omitempty"`
}
```

## Error Handling

The FlowEngine interface uses standard Go error types with specific error types for common conditions:

```go
var (
    ErrFlowNotFound     = errors.New("flow not found")
    ErrFlowExists       = errors.New("flow already exists")
    ErrInvalidConfig    = errors.New("invalid flow configuration")
    ErrProcessNotFound  = errors.New("process type not found")
    ErrInvalidState     = errors.New("invalid flow state for operation")
)
```

## Implementation Requirements

1. **Thread Safety**
   - All methods must be safe for concurrent access
   - Internal state must be protected with appropriate synchronization

2. **Context Support**
   - Future versions may add context.Context parameters
   - Implementations should consider cancellation support

3. **Validation**
   - CreateFlow must validate complete configuration
   - Edge connections must verify port compatibility
   - Process types must be verified to exist

4. **Resource Management**
   - Stopped flows should clean up all resources
   - Long-running operations should support cancellation
   - Memory leaks must be prevented

## Usage Examples

### Creating and Starting a Flow

```go
engine := NewFlowEngine()

config := FlowConfig{
    ID: "example-flow",
    Nodes: map[string]NodeConfig{
        "reader": {
            Type: "FileReader",
            Config: map[string]any{
                "filename": "input.txt",
            },
        },
        "writer": {
            Type: "FileWriter",
            Config: map[string]any{
                "filename": "output.txt",
            },
        },
    },
    Edges: []EdgeConfig{
        {
            From: NodePort{Node: "reader", Port: "out"},
            To:   NodePort{Node: "writer", Port: "in"},
        },
    },
}

id, err := engine.CreateFlow(config)
if err != nil {
    log.Fatalf("Failed to create flow: %v", err)
}

if err := engine.StartFlow(id); err != nil {
    log.Fatalf("Failed to start flow: %v", err)
}
```

### Monitoring Flow Status

```go
status, err := engine.GetFlowStatus(id)
if err != nil {
    log.Fatalf("Failed to get status: %v", err)
}

for nodeName, nodeStatus := range status.NodeStatus {
    fmt.Printf("Node %s: %s (%d messages)\n",
        nodeName,
        nodeStatus.State,
        nodeStatus.MessagesProcessed)
}
```