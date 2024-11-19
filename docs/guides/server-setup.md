# Server Setup Guide

This guide explains how to set up and run the noPromises server with documentation support.

## Configuration

The server can be configured through command-line flags:

```bash
# Start server with default settings
go run cmd/server/main.go

# Start with custom port and docs path
go run cmd/server/main.go -port 3000 -docs ./custom-docs
```

### Configuration Options
- `-port`: Server port (default: 8080)
- `-docs`: Documentation files path (default: ./docs)

## Documentation Access

Once the server is running, documentation is available at:

- Documentation: `http://localhost:8080/docs`
- Network Diagrams: `http://localhost:8080/diagrams/network/{id}`
- API Documentation: `http://localhost:8080/api-docs`

## Using Make Commands

The Makefile provides convenient commands for server management:

```bash
# Start server with default settings
make server-start

# Start on custom port
make server-start-port-3000

# Stop server
make server-stop
```

## Directory Structure

The documentation system expects the following structure:
```
docs/
├── api/              # API documentation
├── architecture/     # Architecture documentation
├── guides/          # User guides
└── README.md        # Documentation home
```

## Adding Documentation

1. Create Markdown files in appropriate directories
2. Use relative links for navigation
3. Include code examples with syntax highlighting
4. Add diagrams using Mermaid syntax

## Network Visualization

To view network diagrams:

1. Create a flow through the API
2. Access `http://localhost:8080/diagrams/network/{flow-id}`
3. For live updates: `http://localhost:8080/diagrams/network/{flow-id}/live`
