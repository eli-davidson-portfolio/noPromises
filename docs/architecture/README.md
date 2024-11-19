# Architecture Documentation

## Overview
This documentation describes the architecture of the noPromises Flow-Based Programming system.

## System Components

### Core Server
- HTTP server implementation
- Flow management
- Process registry
- Request handling
- Error management

### Documentation Server
- Markdown rendering
- API documentation (Swagger/OpenAPI)
- Network visualization
- Live updates
- Static file serving

### Network Visualization
- Mermaid diagram generation
- Real-time state updates
- WebSocket integration
- Interactive diagrams

## Document Structure

### Subsystems (`/subsystems`)
- `server.md` - Core server architecture
- `docs-server.md` - Documentation server
- `server-config.md` - Configuration system
- `server-engine.md` - Flow engine interface
- `server-api.md` - API subsystem

### Patterns (`/patterns`)
- `server-patterns.md` - Server design patterns
- `api-patterns.md` - API design patterns
- `network-patterns.md` - Network topology patterns
- `component-patterns.md` - Component implementation

### Testing (`/testing`)
- `server-testing.md` - Server test patterns
- `network-testing.md` - Network test strategies
- `component-testing.md` - Component testing
- `performance-testing.md` - Performance testing

### Observability (`/observability`)
- `metrics.md` - Metrics collection
- `tracing.md` - Distributed tracing
- `visualization/` - Network visualization
  - `network-diagrams.md` - Mermaid diagrams

### Security (`/security`)
- `access-control.md` - Access control
- `server-auth.md` - Authentication

## Key Architectural Concepts

### Flow-Based Programming
- Independent processes
- Message passing
- Port-based communication
- Network topology

### Documentation First
- Integrated documentation server
- Live API documentation
- Network visualization
- Real-time updates

### Type Safety
- Strongly typed messages
- Port type validation
- Configuration validation
- Error handling

### Concurrency
- Go routines for processes
- Channel-based communication
- Context cancellation
- Resource cleanup

## Integration Points

### Main Server Integration
```go
type Server struct {
    docs     *docs.Server    // Documentation server
    router   *mux.Router     // Main router
    flows    *FlowManager    // Flow management
    engine   FlowEngine      // Flow execution
}
```

### Documentation Integration
```go
func (s *Server) setupDocs() {
    s.router.Mount("/docs", s.docs.Router())
    s.router.Mount("/api-docs", s.docs.SwaggerUI())
    s.router.Mount("/diagrams", s.docs.Diagrams())
}
```

## Documentation Conventions

### Code Examples
- Include language identifiers
- Show complete, working examples
- Add explanatory comments
- Include error handling

### Diagrams
- Use Mermaid syntax
- Follow left-to-right flow
- Include state information
- Label all connections

### API Documentation
- Follow OpenAPI 3.0
- Include request/response examples
- Document all status codes
- Specify content types

### Architecture Diagrams
- Show component relationships
- Include data flow
- Mark integration points
- Note security boundaries