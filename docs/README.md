# noPromises Documentation

Welcome to the noPromises documentation. This is a Flow-Based Programming (FBP) implementation in Go.

## Core Components

### Server
- HTTP server implementation
- Flow management
- Process registry
- Documentation serving
- API endpoints

### Documentation System
- Markdown rendering
- API documentation (Swagger/OpenAPI)
- Network visualization
- Live updates

### Flow Management
- Flow creation and control
- Process type registry
- Network orchestration
- State management

## Getting Started

### Installation
```bash
go get github.com/elleshadow/noPromises
```

### Basic Usage
```go
// Create server
srv, err := server.NewServer(server.Config{
    Port: 8080,
    DocsPath: "./docs",
})
if err != nil {
    log.Fatal(err)
}

// Register process types
srv.RegisterProcessType("FileReader", &FileReaderFactory{})

// Start server
if err := srv.Start(context.Background()); err != nil {
    log.Fatal(err)
}
```

## Documentation Sections

### Architecture
- [Core Concepts](architecture/core-concepts.md)
- [Design Principles](architecture/design-principles.md)
- [Error Handling](architecture/error-handling.md)
- [Network Orchestration](architecture/network-orchestration.md)

### API
- [Endpoints](api/endpoints.md)
- [Schemas](api/schemas.md)
- [OpenAPI Spec](api/swagger.json)

### Guides
- [Getting Started](guides/getting-started.md)
- [Best Practices](guides/best-practices.md)
- [Advanced Usage](guides/advanced-usage.md)

## Current Status

### Implemented ✅
- Basic server implementation
- Flow management
- Process registry
- Documentation server
- Network visualization
- HTML documentation rendering
- Swagger UI integration
- Mermaid diagram generation

### In Progress 🚧
- WebSocket implementation
- Process creation
- Server integration tests
- Advanced error handling
- Performance optimizations
