# noPromises: Classical Flow-Based Programming in Go

noPromises is a strict implementation of J. Paul Morrison's Flow-Based Programming (FBP) paradigm in Go. It leverages Go's channel-based concurrency and type system to create truly independent processes that communicate solely through message passing.

## Features

### Core FBP Implementation
- Independent processes
- Message passing via channels
- Port-based communication
- Network-based topology
- Strong typing

### Database System
- SQLite with WAL mode
- Migration management
- Version tracking
- Transaction safety
- Foreign key support

### Documentation System
- Markdown with HTML rendering
- API documentation (Swagger/OpenAPI)
- Network visualization
- Live updates
- Interactive diagrams

### Server Components
- RESTful API
- Flow management
- Process registry
- WebSocket support
- Error handling

## Getting Started

1. Prerequisites
```bash
# Ensure Go 1.21+ is installed
go version

# Install required tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

2. Installation
```bash
git clone https://github.com/elleshadow/noPromises
cd noPromises
make install-hooks
```

3. Start Server
```bash
# Build and start on port 8080
make server-build
make server-start

# Or custom port with docs
make server-start-port-3000 DOCS_PATH=./docs
```

## Documentation Access

Once running, access documentation at:

```bash
# Main documentation
open http://localhost:8080/docs

# API documentation
open http://localhost:8080/api-docs

# Network visualization
open http://localhost:8080/diagrams/network/{flow-id}

# Live updates
open http://localhost:8080/diagrams/network/{flow-id}/live
```

## Documentation Structure
```
docs/
├── api/              # API documentation
│   ├── endpoints.md  # API endpoint details
│   └── swagger.json  # OpenAPI specification
├── architecture/     # Architecture documentation
│   ├── patterns/    # Design patterns
│   ├── subsystems/  # Subsystem details
│   └── README.md    # Architecture overview
├── guides/          # User guides
└── README.md        # Documentation home
```

## Development

### Make Commands
| Command | Description |
|---------|-------------|
| `make all` | Run all checks and build |
| `make check` | Run linter and tests |
| `make test` | Run tests with race detection |
| `make build` | Build all binaries |
| `make server-start` | Start server on port 8080 |
| `make server-stop` | Stop running server |
| `make new-migration name="migration_name"` | Create new migration files |

### Database Migrations
```bash
# Create new migration files
make new-migration name="create_users_table"

# This creates:
# - internal/db/migrations/NNNNNN_create_users_table.up.sql
# - internal/db/migrations/NNNNNN_create_users_table.down.sql
```

Migration files follow the format:
- `{version}_{name}.up.sql` for forward migrations
- `{version}_{name}.down.sql` for rollback migrations
- Version numbers are 6-digit sequential numbers
- Names should be descriptive and use underscores

### Testing
```bash
# Run all tests
make test

# Run with race detection
go test -race ./...

# Run specific tests
go test ./pkg/server/...
```

### Documentation Updates
1. Edit Markdown files in `docs/`
2. Files are served directly
3. Support for Mermaid diagrams
4. Live updates for network diagrams

## Current Status

### Implemented ✅
- Basic server implementation
- Flow management
- Process registry
- RESTful API
- Documentation server
- Network visualization
- HTML documentation rendering
- Swagger UI integration
- Mermaid diagram generation
- Database migration system
- SQLite with WAL mode
- Transaction-safe migrations

### Coming Soon 🚧
- WebSocket implementation
- Authentication system
- Advanced error handling
- Performance optimizations
- Additional process types
- Monitoring system
- Live documentation updates

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## License

MIT License - See [LICENSE](LICENSE) for details

