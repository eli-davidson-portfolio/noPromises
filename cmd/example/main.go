package main

import (
	"context"
	"log"

	"github.com/elleshadow/noPromises/pkg/server"
)

func main() {
	// Create new server
	srv, err := server.NewServer(server.Config{
		Port:     8080,
		DocsPath: "./docs",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Register a test process type
	srv.RegisterProcessType("FileReader", &MockFileReaderFactory{})

	// Start the server
	log.Printf("Starting server on port 8080...")
	if err := srv.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}

// Mock process factory for testing
type MockFileReaderFactory struct{}

func (f *MockFileReaderFactory) Create(_ map[string]interface{}) (server.Process, error) {
	return &MockFileReader{
		id: "mock-file-reader",
	}, nil
}

type MockFileReader struct {
	id string
}

func (m *MockFileReader) Start(_ context.Context) error { return nil }
func (m *MockFileReader) Stop(_ context.Context) error  { return nil }
func (m *MockFileReader) ID() string                    { return m.id }
