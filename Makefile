# Build directories
BUILD_DIR := bin
WEB_DIR := web

.PHONY: all install-hooks check lint test build clean server-build server-start server-start-port-% server-stop copy-web-assets test-web new-migration

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
	go build -o $(BUILD_DIR)/nop cmd/nop/main.go

# Build the server binary
server-build: copy-web-assets
	go build -o $(BUILD_DIR)/server cmd/server/main.go

# Start the server (default port 8080)
server-start: server-build
	@if [ -f .server.pid ]; then \
		echo "Server already running. Use 'make server-stop' first."; \
		exit 1; \
	fi
	@echo "Starting server on port 8080..."
	@$(BUILD_DIR)/server -port 8080 & echo $$! > .server.pid
	@echo "Server started (PID: $$(cat .server.pid))"

# Start server on custom port
server-start-port-%: server-build
	@if [ -f .server.pid ]; then \
		echo "Server already running. Use 'make server-stop' first."; \
		exit 1; \
	fi
	@echo "Starting server on port $*..."
	@$(BUILD_DIR)/server -port $* & echo $$! > .server.pid
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
	rm -rf $(BUILD_DIR)/
	rm -f .server.pid

# Setup development environment
setup: install-hooks
	go mod download
	go mod verify

# Copy web assets
copy-web-assets:
	@echo "Copying web assets..."
	@mkdir -p $(BUILD_DIR)/web
	@if [ -d "$(WEB_DIR)" ]; then \
		cp -r $(WEB_DIR)/* $(BUILD_DIR)/web/ 2>/dev/null || true; \
	else \
		echo "Warning: $(WEB_DIR) directory not found"; \
		mkdir -p $(WEB_DIR)/templates $(WEB_DIR)/static/css $(WEB_DIR)/static/js; \
		touch $(WEB_DIR)/templates/index.html; \
		touch $(WEB_DIR)/static/css/style.css; \
		touch $(WEB_DIR)/static/js/main.js; \
	fi

# Test web interface
test-web: 
	@echo "Running web interface tests..."
	go test -v -race ./internal/server/web/...

# Create new migration files
new-migration:
	@if [ -z "$(name)" ]; then \
		echo "Error: Missing migration name. Usage: make new-migration name=<migration_name>"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(name)"
	@next_version=$$(printf "%06d" $$(( $$(ls internal/db/migrations/*.up.sql 2>/dev/null | wc -l) + 1 ))); \
	up_file="internal/db/migrations/$${next_version}_$$(echo $(name) | tr ' ' '_').up.sql"; \
	down_file="internal/db/migrations/$${next_version}_$$(echo $(name) | tr ' ' '_').down.sql"; \
	touch "$$up_file" "$$down_file"; \
	echo "-- Migration: $(name)" > "$$up_file"; \
	echo "-- Up migration" >> "$$up_file"; \
	echo "\n" >> "$$up_file"; \
	echo "-- Migration: $(name)" > "$$down_file"; \
	echo "-- Down migration" >> "$$down_file"; \
	echo "\n" >> "$$down_file"; \
	echo "Created migration files:"; \
	echo "  $$up_file"; \
	echo "  $$down_file"