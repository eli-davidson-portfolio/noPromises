# Flow Server Subsystem

The server subsystem provides HTTP-based management of FBP networks.

## Core Components

### Flow Server

```go
type FlowServer struct {
    networks map[string]*Network    // Active networks
    mutex    sync.RWMutex          // Thread safety
    router   *chi.Router           // HTTP routing
    port     int                   // Server port
    registry *ProcessRegistry      // Available process types
}
```

### Process Registry

```go
type ProcessRegistry struct {
    processes map[string]ProcessFactory
    mutex     sync.RWMutex
}

type ProcessFactory func(config ProcessConfig) (Process, error)
```

### Flow Configuration

```go
type FlowConfig struct {
    ID      string                 `json:"id"`
    Nodes   map[string]NodeConfig  `json:"nodes"`
    Edges   []EdgeConfig          `json:"edges"`
}

type NodeConfig struct {
    Type   string                 `json:"type"`
    Config map[string]interface{} `json:"config"`
}

type EdgeConfig struct {
    FromNode string `json:"fromNode"`
    FromPort string `json:"fromPort"`
    ToNode   string `json:"toNode"`
    ToPort   string `json:"toPort"`
}
```

## HTTP API

### Flow Management

#### Create Flow
```http
POST /api/flows
Content-Type: application/json

{
    "id": "flow1",
    "nodes": {
        "reader": {
            "type": "FileReader",
            "config": {
                "path": "/path/to/file"
            }
        },
        "transformer": {
            "type": "UpperCase",
            "config": {}
        }
    },
    "edges": [
        {
            "fromNode": "reader",
            "fromPort": "out",
            "toNode": "transformer",
            "toPort": "in"
        }
    ]
}
```

#### List Flows
```http
GET /api/flows

Response:
{
    "flows": [
        {
            "id": "flow1",
            "status": "running",
            "nodes": ["reader", "transformer"]
        }
    ]
}
```

#### Get Flow Details
```http
GET /api/flows/{id}

Response:
{
    "id": "flow1",
    "status": "running",
    "nodes": {
        "reader": {
            "type": "FileReader",
            "status": "running",
            "metrics": {
                "packetsProcessed": 100,
                "errors": 0
            }
        }
    },
    "edges": [
        {
            "fromNode": "reader",
            "fromPort": "out",
            "toNode": "transformer",
            "toPort": "in",
            "metrics": {
                "packetsTransferred": 100,
                "bufferUsage": 0.5
            }
        }
    ]
}
```

### Flow Control

#### Start Flow
```http
POST /api/flows/{id}/start
```

#### Stop Flow
```http
POST /api/flows/{id}/stop
```

### Process Registry

#### List Available Processes
```http
GET /api/processes

Response:
{
    "processes": [
        {
            "type": "FileReader",
            "description": "Reads files line by line",
            "ports": {
                "in": [],
                "out": [
                    {
                        "name": "out",
                        "type": "string",
                        "description": "File contents"
                    }
                ]
            },
            "config": {
                "path": {
                    "type": "string",
                    "required": true,
                    "description": "File path to read"
                }
            }
        }
    ]
}
```

## Implementation

### Server Setup
```go
func NewFlowServer(config ServerConfig) *FlowServer {
    r := chi.NewRouter()
    
    server := &FlowServer{
        networks: make(map[string]*Network),
        router:   r,
        port:     config.Port,
        registry: NewProcessRegistry(),
    }
    
    server.setupRoutes()
    return server
}
```

### Flow Creation
```go
func (s *FlowServer) CreateFlow(config FlowConfig) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    network := NewNetwork()
    
    // Create nodes
    for id, nodeConfig := range config.Nodes {
        factory, exists := s.registry.Get(nodeConfig.Type)
        if !exists {
            return fmt.Errorf("unknown node type: %s", nodeConfig.Type)
        }
        
        process, err := factory(nodeConfig.Config)
        if err != nil {
            return fmt.Errorf("creating node %s: %w", id, err)
        }
        
        network.AddProcess(id, process)
    }
    
    // Create connections
    for _, edge := range config.Edges {
        err := network.Connect(
            edge.FromNode, edge.FromPort,
            edge.ToNode, edge.ToPort,
        )
        if err != nil {
            return fmt.Errorf("connecting edge: %w", err)
        }
    }
    
    s.networks[config.ID] = network
    return nil
}
```

## Best Practices

### Configuration Validation
- Validate all flow configurations
- Check for cycles in connections
- Verify required ports are connected
- Validate process configurations

### Error Handling
- Return appropriate HTTP status codes
- Provide detailed error messages
- Clean up on partial failures
- Handle concurrent modifications

### Security
- Validate input data
- Sanitize configurations
- Implement authentication
- Use HTTPS

### Monitoring
- Track network status
- Monitor process metrics
- Log flow changes
- Collect performance data

## Example Usage

### Creating and Starting a Flow
```bash
# Create flow
curl -X POST http://localhost:8080/api/flows \
    -H "Content-Type: application/json" \
    -d @flow-config.json

# Start flow
curl -X POST http://localhost:8080/api/flows/flow1/start

# Monitor flow
curl http://localhost:8080/api/flows/flow1
``` 