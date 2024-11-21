# Subsystem Documentation

This directory contains detailed documentation for each major subsystem in noPromises.

## Core Subsystems

### Server Components
- [Server](server.md) - Core server implementation
- [Server Engine](server-engine.md) - Flow execution engine
- [Server Config](server-config.md) - Configuration management
- [Server API](server-api.md) - API implementation
- [Documentation Server](docs-server.md) - Documentation serving

### Flow Components
- [Process Management](process.md) - Process lifecycle and management
- [Ports](ports.md) - Port implementation and management
- [IPs & IIPs](iips.md) - Information Packets
- [Scheduler](scheduler.md) - Process scheduling and orchestration

## Integration Points

Each subsystem document includes:
- Component overview
- Interface definitions
- Implementation details
- Integration examples
- Best practices
- Testing strategies

## Best Practices

When working with subsystems:
1. Follow established interfaces
2. Maintain proper error handling
3. Include appropriate logging
4. Add relevant metrics
5. Write comprehensive tests

## Related Documentation

- [Design Patterns](../patterns/README.md)
- [Observability](../observability/README.md)
- [API Documentation](/docs/api/README.md) 