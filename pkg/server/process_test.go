package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testProcess struct {
	id    string
	state string
}

func (p *testProcess) Start(_ context.Context) error {
	p.state = "running"
	return nil
}

func (p *testProcess) Stop(_ context.Context) error {
	p.state = "stopped"
	return nil
}

func (p *testProcess) ID() string {
	return p.id
}

type testProcessFactory struct{}

func (f *testProcessFactory) Create(_ map[string]interface{}) (Process, error) {
	return &testProcess{
		id:    "test-process",
		state: "created",
	}, nil
}

func TestProcessRegistry(t *testing.T) {
	registry := &ProcessRegistry{
		processes: make(map[string]ProcessFactory),
	}

	factory := &testProcessFactory{}
	registry.Register("test", factory)

	// Test retrieval
	got, exists := registry.Get("test")
	require.True(t, exists)
	assert.Equal(t, factory, got)

	// Test non-existent type
	_, exists = registry.Get("unknown")
	assert.False(t, exists)

	// Test process creation
	proc, err := factory.Create(nil)
	require.NoError(t, err)
	assert.Equal(t, "test-process", proc.ID())
}
