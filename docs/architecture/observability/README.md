# Observability Architecture

This directory contains the architectural documentation for the observability system, which provides comprehensive monitoring, metrics collection, tracing, and visualization capabilities.

## Core Components

### Metrics Collection
- [metrics.md](./metrics.md) - Core metrics architecture including process and network metrics
- [server-metrics.md](./server-metrics.md) - Server-specific metrics implementation including HTTP and resource metrics

### Monitoring
- [monitoring.md](./monitoring.md) - Health monitoring and alerting system architecture
- Components include:
  - Health monitoring
  - Alert management
  - Process and network health checks
  - System diagnostics
  - Performance profiling

### Distributed Tracing
- [tracing.md](./tracing.md) - Distributed tracing implementation
- Features include:
  - Trace context management
  - Span tracking
  - Sampling strategies
  - OpenTelemetry integration
  - Performance optimization

### Visualization
- [visualization.md](./visualization.md) - Network visualization architecture
- [visualization/network-diagrams.md](./visualization/network-diagrams.md) - Network diagram generation using Mermaid
- Key capabilities:
  - Real-time network visualization
  - Interactive network navigation
  - Process inspection
  - Performance visualization
  - WebSocket-based live updates

## Integration Points

The observability system integrates with:
- Prometheus for metrics collection and storage
- OpenTelemetry for distributed tracing
- WebSocket endpoints for real-time visualization
- HTTP API endpoints for data access
- Process and network components for data collection

## Best Practices

Each component includes best practices for:
- Performance optimization
- Resource management
- Testing strategies
- Integration patterns
- User experience considerations

## Getting Started

1. Review the metrics architecture to understand data collection
2. Set up monitoring and alerting based on monitoring.md
3. Implement distributed tracing following tracing.md
4. Configure visualization components using visualization.md

## Directory Structure

```
observability/
├── README.md
├── metrics.md
├── monitoring.md
├── server-metrics.md
├── tracing.md
├── visualization.md
└── visualization/
    └── network-diagrams.md
``` 