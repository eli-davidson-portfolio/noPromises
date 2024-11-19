# Getting Started with noPromises

This guide will help you get started with noPromises.

## Installation

1. Prerequisites
```bash
# Ensure Go 1.21+ is installed
go version

# Install required tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

2. Clone the Repository
```bash
git clone https://github.com/elleshadow/noPromises
cd noPromises
```

3. Install Development Tools
```bash
make install-hooks
```

## Running the Server

1. Build the Server
```bash
make server-build
```

2. Start the Server
```bash
# Start on default port 8080
make server-start

# Or start on custom port with docs
make server-start-port-3000 DOCS_PATH=./docs
```

## Accessing Documentation

Once the server is running, you can access:

1. Main Documentation
```bash
open http://localhost:8080/docs
```

2. API Documentation
```bash
open http://localhost:8080/api-docs
```

3. Network Visualization
```bash
# View specific network diagram
open http://localhost:8080/diagrams/network/{flow-id}

# View live updates
open http://localhost:8080/diagrams/network/{flow-id}/live
```

## Creating Your First Flow

1. Create a Flow Configuration
```json
{
    "id": "example-flow",
    "nodes": {
        "reader": {
            "type": "FileReader",
            "config": {
                "filename": "input.txt"
            }
        },
        "writer": {
            "type": "FileWriter",
            "config": {
                "filename": "output.txt"
            }
        }
    },
    "edges": [
        {
            "fromNode": "reader",
            "fromPort": "out",
            "toNode": "writer",
            "toPort": "in"
        }
    ]
}
```

2. Create the Flow
```bash
curl -X POST http://localhost:8080/api/v1/flows \
    -H "Content-Type: application/json" \
    -d @flow.json
```

3. Start the Flow
```bash
curl -X POST http://localhost:8080/api/v1/flows/example-flow/start
```

4. Monitor the Flow
```bash
# View flow status
curl http://localhost:8080/api/v1/flows/example-flow/status

# View flow visualization
open http://localhost:8080/diagrams/network/example-flow
```

## Development Workflow

1. Make Changes
- Edit code
- Add tests
- Update documentation

2. Run Checks
```bash
# Run all checks
make check

# Run specific checks
make lint
make test
make format-check
```

3. Build and Test
```bash
# Build all
make build

# Run tests with race detection
make test
```

## Next Steps

1. Review the [Architecture Documentation](../architecture/README.md)
2. Explore [API Documentation](../api/endpoints.md)
3. Read [Best Practices](best-practices.md)
4. Check [Contributing Guidelines](../CONTRIBUTING.md)
