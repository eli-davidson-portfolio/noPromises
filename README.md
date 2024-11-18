# noPromises: Classical Flow-Based Programming in Go

noPromises is a strict implementation of J. Paul Morrison's Flow-Based Programming (FBP) paradigm in Go. It leverages Go's channel-based concurrency and type system to create truly independent processes that communicate solely through message passing.

## Core Components

### 1. Information Packets (IPs)
```go
type IP[T any] struct {
    Data     T
    Metadata map[string]any
}
```
- Type-safe data transport
- Metadata support
- Thread-safe operations

### 2. Ports
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
- Type-safe connections
- Connection limits
- Buffered channels
- Fan-out support

### 3. Processes
```go
type Process interface {
    Initialize(ctx context.Context) error
    Process(ctx context.Context) error
    Shutdown(ctx context.Context) error
    IsInitialized() bool
}
```
- Context-aware lifecycle
- Clean initialization/shutdown
- State management
- Error propagation

### 4. Networks
```go
type Network struct {
    processes map[string]Process
}
```
- Process management
- Connection orchestration
- Error handling
- Clean shutdown

## Example Usage

### Creating a Process
```go
type CustomProcess struct {
    process.BaseProcess
    in  *ports.Port[string]
    out *ports.Port[string]
}

func (p *CustomProcess) Process(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            packet, err := p.in.Receive(ctx)
            if err != nil {
                return err
            }
            // Process data...
            result := ip.New[string](processedData)
            if err := p.out.Send(ctx, result); err != nil {
                return err
            }
        }
    }
}
```

### Building a Network
```go
net := network.New()

// Add processes
net.AddProcess("proc1", NewProcess1())
net.AddProcess("proc2", NewProcess2())

// Connect processes
net.Connect("proc1", "out", "proc2", "in")

// Run network
ctx := context.Background()
if err := net.Run(ctx); err != nil {
    // Handle error
}
```

## Server Components (Coming Soon)

### Flow Server
```go
type FlowServer struct {
    networks map[string]*Network    // Network management
    registry *ProcessRegistry       // Process type registry
    router   *chi.Router           // HTTP routing
}
```

### Process Registry
```go
type ProcessRegistry struct {
    processes map[string]ProcessFactory
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
```

### HTTP API (Planned)
- `POST /api/flows` - Create flow
- `GET /api/flows` - List flows
- `GET /api/flows/{id}` - Get flow details
- `DELETE /api/flows/{id}` - Delete flow
- `POST /api/flows/{id}/start` - Start flow
- `POST /api/flows/{id}/stop` - Stop flow
- `GET /api/processes` - List available processes

## Current Status

### Implemented
- ✅ Core IP system
- ✅ Port management
- ✅ Process lifecycle
- ✅ Network orchestration
- ✅ Basic error handling
- ✅ Context support
- ✅ Type safety

### In Progress
- 🚧 Flow server implementation
- 🚧 Process registry
- 🚧 HTTP API
- 🚧 Flow configuration
- 🚧 Advanced error handling
- 🚧 Monitoring system
- 🚧 Performance optimizations

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

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## License

MIT License - See [LICENSE](LICENSE) for details
