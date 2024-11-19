.PHONY: all install-hooks check lint test build clean server-build server-start server-start-port-% server-stop

# Default target runs all checks and builds
all: check build

# Install git hooks
install-hooks:
	@chmod +x scripts/install-hooks.sh
	@./scripts/install-hooks.sh

# Run all checks (same as pre-commit hook)
check: lint test format-check

# Run linter
lint:
	golangci-lint run

# Run tests with race detection
test:
	go test -v -race ./...

# Check code formatting
format-check:
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted:"; \
		gofmt -l .; \
		exit 1; \
	fi

# Format code
format:
	gofmt -w .

# Build all binaries
build: server-build
	go build -o bin/nop cmd/nop/main.go

# Build the server binary
server-build:
	go build -o bin/server cmd/server/main.go

# Start the server (default port 8080)
server-start: server-build
	@if [ -f .server.pid ]; then \
		echo "Server already running. Use 'make server-stop' first."; \
		exit 1; \
	fi
	@echo "Starting server on port 8080..."
	@bin/server -port 8080 & echo $$! > .server.pid
	@echo "Server started (PID: $$(cat .server.pid))"

# Start server on custom port
server-start-port-%: server-build
	@if [ -f .server.pid ]; then \
		echo "Server already running. Use 'make server-stop' first."; \
		exit 1; \
	fi
	@echo "Starting server on port $*..."
	@bin/server -port $* & echo $$! > .server.pid
	@echo "Server started (PID: $$(cat .server.pid))"

# Stop the server
server-stop:
	@if [ -f .server.pid ]; then \
		echo "Stopping server (PID: $$(cat .server.pid))..."; \
		kill $$(cat .server.pid) || true; \
		rm .server.pid; \
		echo "Server stopped"; \
	else \
		echo "No server running"; \
	fi

# Clean build artifacts
clean: server-stop
	go clean
	rm -f coverage.out
	rm -rf bin/
	rm -f .server.pid

# Setup development environment
setup: install-hooks
	go mod download
	go mod verify