package server_test

import (
	"context"
	"testing"

	"github.com/elleshadow/noPromises/pkg/server"
)

//nolint:unused // Will be used when tests are implemented
type mockProcess struct {
	id     string
	status string
}

//nolint:unused // Will be used when tests are implemented
func (m *mockProcess) Start(_ context.Context) error {
	m.status = "running"
	return nil
}

//nolint:unused // Will be used when tests are implemented
func (m *mockProcess) Stop(_ context.Context) error {
	m.status = "stopped"
	return nil
}

//nolint:unused // Will be used when tests are implemented
func (m *mockProcess) ID() string {
	return m.id
}

//nolint:unused // Will be used when tests are implemented
type mockProcessFactory struct{}

//nolint:unused // Will be used when tests are implemented
func (f *mockProcessFactory) Create(_ map[string]interface{}) (server.Process, error) {
	return &mockProcess{
		id:     "test-process",
		status: "created",
	}, nil
}

func TestProcesses(t *testing.T) {
	t.Skip("Process creation not implemented yet")
}
