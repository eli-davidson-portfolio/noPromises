.PHONY: test lint clean build

test:
    go test -v -race ./...

lint:
    golangci-lint run

build:
    go build -v ./...

clean:
    go clean
    rm -f coverage.out