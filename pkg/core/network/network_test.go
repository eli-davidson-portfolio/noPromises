package network_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/core/network"
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
			result := ip.New[string](strings.ToUpper(packet.Data()))

			err = p.out.Send(ctx, result)
			if err != nil {
				return err
			}
		}
	}
}

// FailingProcess is a process that always returns an error
type FailingProcess struct {
	process.BaseProcess
}

func (p *FailingProcess) Process(_ context.Context) error {
	return errors.New("simulated failure")
}

func TestNetwork(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		t.Run("empty network", func(t *testing.T) {
			net := network.New()
			assert.NotNil(t, net)
			assert.Equal(t, 0, net.ProcessCount())
		})

		t.Run("with processes", func(t *testing.T) {
			net := network.New()

			err := net.AddProcess("proc1", NewTestProcess())
			require.NoError(t, err)

			err = net.AddProcess("proc2", NewTestProcess())
			require.NoError(t, err)

			assert.Equal(t, 2, net.ProcessCount())
		})
	})

	t.Run("connections", func(t *testing.T) {
		t.Run("valid connection", func(t *testing.T) {
			net := network.New()

			proc1 := NewTestProcess()
			proc2 := NewTestProcess()

			require.NoError(t, net.AddProcess("proc1", proc1))
			require.NoError(t, net.AddProcess("proc2", proc2))

			err := net.Connect("proc1", "out", "proc2", "in")
			assert.NoError(t, err)
		})

		t.Run("invalid process", func(t *testing.T) {
			net := network.New()

			proc1 := NewTestProcess()
			require.NoError(t, net.AddProcess("proc1", proc1))

			err := net.Connect("proc1", "out", "nonexistent", "in")
			assert.Error(t, err)
		})

		t.Run("invalid port", func(t *testing.T) {
			net := network.New()

			proc1 := NewTestProcess()
			proc2 := NewTestProcess()

			require.NoError(t, net.AddProcess("proc1", proc1))
			require.NoError(t, net.AddProcess("proc2", proc2))

			err := net.Connect("proc1", "nonexistent", "proc2", "in")
			assert.Error(t, err)
		})
	})

	t.Run("execution", func(t *testing.T) {
		t.Run("successful execution", func(t *testing.T) {
			net := network.New()

			// Add processes
			proc1 := NewTestProcess()
			proc2 := NewTestProcess()

			require.NoError(t, net.AddProcess("proc1", proc1))
			require.NoError(t, net.AddProcess("proc2", proc2))

			// Connect processes
			require.NoError(t, net.Connect("proc1", "out", "proc2", "in"))

			// Create input and output channels for testing
			inputCh := make(chan *ip.IP[string], 1)
			outputCh := make(chan *ip.IP[string], 1)

			// Connect input and output channels
			require.NoError(t, proc1.in.Connect(inputCh))
			require.NoError(t, proc2.out.Connect(outputCh))

			// Start network in background
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			networkDone := make(chan error, 1)
			go func() {
				networkDone <- net.Run(ctx)
			}()

			// Send test data
			testData := "test"
			inputCh <- ip.New[string](testData)

			// Wait for result with timeout
			select {
			case result := <-outputCh:
				assert.Equal(t, strings.ToUpper(testData), result.Data())
				cancel() // Stop network after successful test
			case err := <-networkDone:
				t.Fatalf("Network stopped unexpectedly: %v", err)
			case <-time.After(time.Second):
				t.Fatal("Test timed out waiting for result")
			}

			// Wait for network to stop
			err := <-networkDone
			assert.ErrorIs(t, err, context.Canceled)
		})

		t.Run("context cancellation", func(t *testing.T) {
			net := network.New()

			proc1 := NewTestProcess()
			require.NoError(t, net.AddProcess("proc1", proc1))

			ctx, cancel := context.WithCancel(context.Background())

			// Start network
			errCh := make(chan error)
			go func() {
				errCh <- net.Run(ctx)
			}()

			// Cancel context
			cancel()

			// Check that network stopped
			err := <-errCh
			assert.Error(t, err)
		})
	})

	t.Run("error handling", func(t *testing.T) {
		t.Run("process error propagation", func(t *testing.T) {
			net := network.New()

			// Add failing process
			failingProc := &FailingProcess{}
			require.NoError(t, net.AddProcess("fail", failingProc))

			// Run network
			err := net.Run(context.Background())
			assert.Error(t, err)
		})

		t.Run("cleanup on error", func(t *testing.T) {
			net := network.New()

			proc1 := NewTestProcess()
			failingProc := &FailingProcess{}

			require.NoError(t, net.AddProcess("proc1", proc1))
			require.NoError(t, net.AddProcess("fail", failingProc))

			// Run network
			err := net.Run(context.Background())
			assert.Error(t, err)

			// Check that all processes are properly shut down
			assert.False(t, proc1.IsInitialized())
		})
	})
}
