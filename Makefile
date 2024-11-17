.PHONY: lint test build

lint:
	golangci-lint run

test:
	go test -v ./...

build:
	go build -o bin/nop cmd/nop/main.go

clean:
	go clean
	rm -f coverage.out