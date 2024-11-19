package debug

import (
	"context"
	"testing"
	"time"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	logger := NewLogger[string]("test")

	// Create test channels
	inCh := make(chan *ip.IP[string], 1)
	outCh := make(chan *ip.IP[string], 1)

	// Connect ports
	require.NoError(t, ports.Connect(logger.InPort, inCh))
	require.NoError(t, ports.Connect(logger.OutPort, outCh))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start processing in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- logger.Process(ctx)
	}()

	// Send test data
	testData := "test message"
	require.NoError(t, logger.InPort.Send(ctx, ip.New(testData)))

	// Verify output
	select {
	case packet := <-outCh:
		assert.Equal(t, testData, packet.Data())
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for output")
	}

	// Verify clean shutdown
	cancel()
	select {
	case err := <-errCh:
		assert.Equal(t, context.Canceled, err)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for shutdown")
	}
}
