# noPromises: Classical Flow-Based Programming in Go

noPromises is a strict implementation of J. Paul Morrison's Flow-Based Programming (FBP) paradigm in Go. It leverages Go's channel-based concurrency and type system to create truly independent processes that communicate solely through message passing.

## Core Components

### 1. Server
```go
type Server struct {
    config    Config
    router    *mux.Router
    flows     *FlowManager
    processes *ProcessRegistry
    Handler   http.Handler
}
```
- RESTful API
- Flow management
- Process registry
- Middleware support

### 2. Flow Management
```go
type FlowManager struct {
    flows map[string]*ManagedFlow
    mu    sync.RWMutex
}

type ManagedFlow struct {
    ID        string
    Config    map[string]interface{}
    State     FlowState
    StartTime *time.Time
    Error     string
}
```
- Flow lifecycle management
- State transitions
- Concurrent access
- Error handling

### 3. Process Registry
```go
type ProcessRegistry struct {
    processes map[string]ProcessFactory
    mu        sync.RWMutex
}

type ProcessFactory interface {
    Create(config map[string]interface{}) (Process, error)
}

type Process interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```
- Process type registration
- Factory pattern
- Configuration validation
- Context-aware lifecycle

## API Endpoints

### Flow Management
- `POST /api/v1/flows` - Create flow
- `GET /api/v1/flows` - List flows
- `GET /api/v1/flows/{id}` - Get flow details
- `DELETE /api/v1/flows/{id}` - Delete flow
- `POST /api/v1/flows/{id}/start` - Start flow
- `POST /api/v1/flows/{id}/stop` - Stop flow
- `GET /api/v1/flows/{id}/status` - Get flow status

### Flow States
- `created`: Initial state after flow creation
- `starting`: Flow is in the process of starting
- `running`: Flow is actively running
- `stopping`: Flow is in the process of stopping
- `stopped`: Flow has been stopped
- `error`: Flow encountered an error

## Example Usage

### Creating a Flow
```http
POST /api/v1/flows
Content-Type: application/json

{
    "id": "example-flow",
    "nodes": {
        "reader": {
            "type": "FileReader",
            "config": {
                "filename": "input.txt"
            }
        }
    },
    "edges": []
}
```

### Starting a Flow
```http
POST /api/v1/flows/example-flow/start
```

### Getting Flow Status
```http
GET /api/v1/flows/example-flow/status

Response:
{
    "data": {
        "id": "example-flow",
        "state": "running",
        "started_at": "2024-01-01T12:00:00Z"
    }
}
```

## Current Status

### Implemented
- ✅ Basic server implementation
- ✅ Flow management
- ✅ Process registry
- ✅ RESTful API
- ✅ Flow lifecycle
- ✅ Error handling
- ✅ Concurrent operations
- ✅ Request validation

### Coming Soon
- 🚧 Core FBP components (IPs, Ports, Networks)
- 🚧 Process implementations
- 🚧 Flow visualization
- 🚧 Monitoring system
- 🚧 Advanced error handling
- 🚧 Performance optimizations
- 🚧 Documentation server
- 🚧 Network visualization

## Development Requirements
- Go 1.21+
- golangci-lint
- make

## Getting Started

1. Clone the repository
```bash
git clone https://github.com/elleshadow/noPromises
```

2. Install dependencies
```bash
make setup
```

3. Run tests
```bash
make test
```

4. Build the server
```bash
make server-build
```

5. Start the server
```bash
# Start on default port 8080
make server-start

# Or start on custom port
make server-start-port-3000
```

6. Stop the server
```bash
make server-stop
```

### Make Commands

| Command | Description |
|---------|-------------|
| `make all` | Run all checks and build |
| `make check` | Run linter and tests |
| `make lint` | Run golangci-lint |
| `make test` | Run tests with race detection |
| `make format` | Format code |
| `make build` | Build all binaries |
| `make server-build` | Build server binary |
| `make server-start` | Start server on port 8080 |
| `make server-start-port-X` | Start server on port X |
| `make server-stop` | Stop running server |
| `make clean` | Clean build artifacts |

### Example API Usage

After starting the server:

1. Create a flow:
```bash
curl -X POST http://localhost:8080/api/v1/flows \
  -H "Content-Type: application/json" \
  -d '{
    "id": "example-flow",
    "nodes": {
      "reader": {
        "type": "FileReader",
        "config": {
          "filename": "input.txt"
        }
      }
    },
    "edges": []
  }'
```

2. Start the flow:
```bash
curl -X POST http://localhost:8080/api/v1/flows/example-flow/start
```

3. Check flow status:
```bash
curl http://localhost:8080/api/v1/flows/example-flow/status
```

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## License

MIT License - See [LICENSE](LICENSE) for details

## Planned Features

### Network Visualization (Coming Soon)
```mermaid
graph LR
    reader[FileReader]:::running
    transform[UpperCase]:::error
    writer[FileWriter]:::waiting
    reader -->|out| transform
    transform -->|out| writer
    
    classDef running fill:#d4edda,stroke:#28a745;
    classDef error fill:#f8d7da,stroke:#dc3545;
    classDef waiting fill:#fff3cd,stroke:#ffc107;
```

### Documentation Server (Planned)
```bash
# Start server
go run cmd/server/main.go

# View documentation (coming soon)
open http://localhost:8080/docs

# View network visualizations (coming soon)
open http://localhost:8080/diagrams/network/flow1

# View API documentation (coming soon)
open http://localhost:8080/api-docs
```

