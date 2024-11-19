package engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockEngine struct {
	started bool
	stopped bool
}

func (m *mockEngine) Start(_ context.Context) error {
	m.started = true
	return nil
}

func (m *mockEngine) Stop(_ context.Context) error {
	m.stopped = true
	return nil
}

func TestEngineLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	engine := &mockEngine{}

	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, engine.started)

	err = engine.Stop(ctx)
	assert.NoError(t, err)
	assert.True(t, engine.stopped)
}
