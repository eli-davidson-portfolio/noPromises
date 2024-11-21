# noPromises Documentation

Welcome to the noPromises documentation. This is a comprehensive guide to the Flow-Based Programming (FBP) implementation in Go.

## Documentation Structure

### [API Documentation](/docs/api)
- [API Overview](api/README.md)
- [Endpoints](api/endpoints.md)
- [JSON Schemas](api/schemas.md)
- [OpenAPI/Swagger](api/swagger.json)

### [Architecture](/docs/architecture)
- [Core Concepts](architecture/core-concepts.md)
- [Design Principles](architecture/design-principles.md)
- [Error Handling](architecture/error-handling.md)
- [Network Orchestration](architecture/network-orchestration.md)

#### [Design Patterns](architecture/patterns)
- [API Patterns](architecture/patterns/api-patterns.md)
- [Component Patterns](architecture/patterns/component-patterns.md)
- [Error Patterns](architecture/patterns/error-patterns.md)
- [Network Patterns](architecture/patterns/network-patterns.md)
- [Server Patterns](architecture/patterns/server-patterns.md)

#### [Observability](architecture/observability)
- [Metrics](architecture/observability/metrics.md)
- [Monitoring](architecture/observability/monitoring.md)
- [Tracing](architecture/observability/tracing.md)
- [Visualization](architecture/observability/visualization.md)

#### [Subsystems](architecture/subsystems)
- [Server](architecture/subsystems/server.md)
- [Server Engine](architecture/subsystems/server-engine.md)
- [Server Config](architecture/subsystems/server-config.md)
- [Server API](architecture/subsystems/server-api.md)
- [Documentation Server](architecture/subsystems/docs-server.md)
- [Process Management](architecture/subsystems/process.md)
- [Ports](architecture/subsystems/ports.md)
- [IPs & IIPs](architecture/subsystems/iips.md)
- [Scheduler](architecture/subsystems/scheduler.md)

### [Guides](/docs/guides)
- Getting Started
- Best Practices
- Advanced Usage
- Troubleshooting

## Contributing

Please read our [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) before submitting contributions.

## Quick Start

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

## Documentation Standards

1. **File Format**
   - Use Markdown for all documentation
   - Include language tags in code blocks
   - Follow consistent heading hierarchy
   - Use relative links for navigation

2. **Code Examples**
   - Include language identifier
   - Provide complete, working examples
   - Add explanatory comments
   - Show both success and error cases

3. **API Documentation**
   - Follow OpenAPI 3.0 specification
   - Include request/response examples
   - Document all status codes
   - Specify content types

4. **Diagrams**
   - Use Mermaid for diagrams
   - Include diagram source
   - Provide diagram description
   - Follow consistent styling

## Getting Help

- Join our Discord server
- Check existing issues
- Review documentation
- Ask in discussions
