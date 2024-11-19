package ports

import (
	"context"
	"testing"
	"time"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPort(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		t.Run("input port", func(t *testing.T) {
			port := NewInput[string]("test", "Test port", true)
			assert.Equal(t, "test", port.Name())
			assert.Equal(t, "Test port", port.Description())
			assert.True(t, port.Required())
			assert.Equal(t, TypeInput, port.Type())
		})

		t.Run("output port", func(t *testing.T) {
			port := NewOutput[string]("test", "Test port", false)
			assert.Equal(t, "test", port.Name())
			assert.Equal(t, "Test port", port.Description())
			assert.False(t, port.Required())
			assert.Equal(t, TypeOutput, port.Type())
		})
	})

	t.Run("connection limits", func(t *testing.T) {
		t.Run("input port limits", func(t *testing.T) {
			port := NewInput[string]("test", "Test port", true)
			port.SetMaxConnections(2)

			ch1 := make(chan *ip.IP[string])
			ch2 := make(chan *ip.IP[string])
			ch3 := make(chan *ip.IP[string])

			err1 := Connect(port, ch1)
			require.NoError(t, err1)

			err2 := Connect(port, ch2)
			require.NoError(t, err2)

			err3 := Connect(port, ch3)
			assert.Error(t, err3)
		})

		t.Run("output port limits", func(t *testing.T) {

			port := NewOutput[string]("test", "Test port", true)
			port.SetMaxConnections(1)

			ch1 := make(chan *ip.IP[string])

			ch2 := make(chan *ip.IP[string])

			err1 := Connect(port, ch1)
			require.NoError(t, err1)

			err2 := Connect(port, ch2)
			assert.Error(t, err2)
		})
	})

	t.Run("send/receive", func(t *testing.T) {
		t.Run("basic send/receive", func(t *testing.T) {
			inPort := NewInput[string]("in", "Input port", true)
			outPort := NewOutput[string]("out", "Output port", true)

			ch := make(chan *ip.IP[string], 1)
			err1 := Connect(inPort, ch)
			require.NoError(t, err1)

			err2 := Connect(outPort, ch)
			require.NoError(t, err2)

			ctx := context.Background()
			testData := "test"
			require.NoError(t, outPort.Send(ctx, ip.New(testData)))

			received, err := inPort.Receive(ctx)
			require.NoError(t, err)
			assert.Equal(t, testData, received.Data())
		})

		t.Run("context cancellation", func(t *testing.T) {
			inPort := NewInput[string]("in", "Input port", true)
			outPort := NewOutput[string]("out", "Output port", true)

			ch := make(chan *ip.IP[string])
			err1 := Connect(inPort, ch)
			require.NoError(t, err1)

			err2 := Connect(outPort, ch)
			require.NoError(t, err2)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := outPort.Send(ctx, ip.New("test"))
			assert.Equal(t, context.Canceled, err)

			_, err = inPort.Receive(ctx)
			assert.Equal(t, context.Canceled, err)
		})

		t.Run("timeout", func(t *testing.T) {
			inPort := NewInput[string]("in", "Input port", true)
			ch := make(chan *ip.IP[string])
			err := Connect(inPort, ch)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			_, err = inPort.Receive(ctx)
			assert.Equal(t, context.DeadlineExceeded, err)
		})
	})
}
