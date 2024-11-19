package control

import (
	"context"
	"testing"
	"time"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelay(t *testing.T) {
	delay := NewDelay[string](100 * time.Millisecond)

	// Create test channels
	inCh := make(chan *ip.IP[string], 1)
	outCh := make(chan *ip.IP[string], 1)

	// Connect ports
	require.NoError(t, ports.Connect(delay.InPort, inCh))
	require.NoError(t, ports.Connect(delay.OutPort, outCh))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start processing in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- delay.Process(ctx)
	}()

	// Send test data and measure delay
	testData := "test message"
	start := time.Now()
	require.NoError(t, delay.InPort.Send(ctx, ip.New(testData)))

	// Verify delayed output
	select {
	case packet := <-outCh:
		elapsed := time.Since(start)
		assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
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
