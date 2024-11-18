.PHONY: all install-hooks check lint test build clean

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

# Build the binary
build:
	go build -o bin/nop cmd/nop/main.go

# Clean build artifacts
clean:
	go clean
	rm -f coverage.out
	rm -rf bin/