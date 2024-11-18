package process_test

import (
	"context"
	"strings"
	"testing"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/core/ports"
	"github.com/elleshadow/noPromises/pkg/core/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProcess is a simple process that transforms strings to uppercase
type TestProcess struct {
	process.BaseProcess
	in  *ports.Port[string]
	out *ports.Port[string]
}

func NewTestProcess() *TestProcess {
	p := &TestProcess{}
	p.in = ports.NewInput[string]("in", "Input port", true)
	p.out = ports.NewOutput[string]("out", "Output port", true)
	return p
}

func (p *TestProcess) GetPort(name string) interface{} {
	switch name {
	case "in":
		return p.in
	case "out":
		return p.out
	default:
		return nil
	}
}

func (p *TestProcess) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			packet, err := p.in.Receive(ctx)
			if err != nil {
				return err
			}

			// Transform data to uppercase
			result := ip.New(strings.ToUpper(packet.Data()))

			err = p.out.Send(ctx, result)
			if err != nil {
				return err
			}
		}
	}
}

func TestProcessLifecycle(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		proc := NewTestProcess()
		ctx := context.Background()

		err := proc.Initialize(ctx)
		require.NoError(t, err)
		assert.True(t, proc.IsInitialized())
	})

	t.Run("shutdown", func(t *testing.T) {
		proc := NewTestProcess()
		ctx := context.Background()

		err := proc.Initialize(ctx)
		require.NoError(t, err)

		err = proc.Shutdown(ctx)
		require.NoError(t, err)
		assert.False(t, proc.IsInitialized())
	})
}

func TestProcessing(t *testing.T) {
	t.Run("basic transformation", func(t *testing.T) {
		proc := NewTestProcess()
		ctx := context.Background()

		inCh := make(chan *ip.IP[string], 1)
		outCh := make(chan *ip.IP[string], 1)

		require.NoError(t, proc.in.Connect(inCh))
		require.NoError(t, proc.out.Connect(outCh))

		// Start processing in background
		go func() {
			err := proc.Process(ctx)
			require.NoError(t, err)
		}()

		// Send test data
		inCh <- ip.New("test")

		// Receive result
		result := <-outCh
		assert.Equal(t, "TEST", result.Data())
	})

	t.Run("context cancellation", func(t *testing.T) {
		proc := NewTestProcess()
		ctx, cancel := context.WithCancel(context.Background())

		inCh := make(chan *ip.IP[string])
		outCh := make(chan *ip.IP[string])

		require.NoError(t, proc.in.Connect(inCh))
		require.NoError(t, proc.out.Connect(outCh))

		// Start processing
		processDone := make(chan error)
		go func() {
			processDone <- proc.Process(ctx)
		}()

		// Cancel context
		cancel()

		// Check that processing stopped
		err := <-processDone
		assert.Error(t, err)
	})
}
