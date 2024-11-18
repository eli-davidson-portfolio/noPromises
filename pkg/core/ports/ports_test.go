package ports_test

import (
	"context"
	"testing"
	"time"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPort(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		t.Run("input port", func(t *testing.T) {
			port := ports.NewInput[string]("test", "Test port", true)
			assert.Equal(t, "test", port.Name())
			assert.Equal(t, "Test port", port.Description())
			assert.True(t, port.Required())
			assert.Equal(t, ports.TypeInput, port.Type())
		})

		t.Run("output port", func(t *testing.T) {
			port := ports.NewOutput[string]("test", "Test port", false)
			assert.Equal(t, "test", port.Name())
			assert.Equal(t, "Test port", port.Description())
			assert.False(t, port.Required())
			assert.Equal(t, ports.TypeOutput, port.Type())
		})
	})

	t.Run("connection limits", func(t *testing.T) {
		t.Run("input port limits", func(t *testing.T) {
			port := ports.NewInput[string]("test", "Test port", true)
			port.SetMaxConnections(2)

			err := port.Connect(make(chan *ip.IP[string]))
			require.NoError(t, err)

			err = port.Connect(make(chan *ip.IP[string]))
			require.NoError(t, err)

			err = port.Connect(make(chan *ip.IP[string]))
			assert.Error(t, err, "should error when exceeding max connections")
		})

		t.Run("output port limits", func(t *testing.T) {
			port := ports.NewOutput[string]("test", "Test port", true)
			port.SetMaxConnections(2)

			err := port.Connect(make(chan *ip.IP[string]))
			require.NoError(t, err)

			err = port.Connect(make(chan *ip.IP[string]))
			require.NoError(t, err)

			err = port.Connect(make(chan *ip.IP[string]))
			assert.Error(t, err, "should error when exceeding max connections")
		})
	})

	t.Run("send/receive", func(t *testing.T) {
		ctx := context.Background()

		t.Run("basic send/receive", func(t *testing.T) {
			in := ports.NewInput[string]("in", "Input port", true)
			out := ports.NewOutput[string]("out", "Output port", true)

			ch := make(chan *ip.IP[string], 1)
			require.NoError(t, in.Connect(ch))
			require.NoError(t, out.Connect(ch))

			packet := ip.New("test")
			err := out.Send(ctx, packet)
			require.NoError(t, err)

			received, err := in.Receive(ctx)
			require.NoError(t, err)
			assert.Equal(t, packet, received)
		})

		t.Run("context cancellation", func(t *testing.T) {
			in := ports.NewInput[string]("in", "Input port", true)
			out := ports.NewOutput[string]("out", "Output port", true)

			ch := make(chan *ip.IP[string])
			require.NoError(t, in.Connect(ch))
			require.NoError(t, out.Connect(ch))

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := out.Send(ctx, ip.New("test"))
			assert.Error(t, err)

			_, err = in.Receive(ctx)
			assert.Error(t, err)
		})

		t.Run("timeout", func(t *testing.T) {
			in := ports.NewInput[string]("in", "Input port", true)
			out := ports.NewOutput[string]("out", "Output port", true)

			ch := make(chan *ip.IP[string])
			require.NoError(t, in.Connect(ch))
			require.NoError(t, out.Connect(ch))

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			err := out.Send(ctx, ip.New("test"))
			assert.Error(t, err)

			_, err = in.Receive(ctx)
			assert.Error(t, err)
		})
	})
}
