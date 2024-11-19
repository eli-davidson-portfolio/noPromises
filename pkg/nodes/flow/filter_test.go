package flow

import (
	"context"
	"testing"
	"time"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	// Create filter that only allows even numbers
	isEven := func(n int) bool { return n%2 == 0 }
	filter := NewFilter[int](isEven)

	// Create test channels
	inCh := make(chan *ip.IP[int], 1)
	outCh := make(chan *ip.IP[int], 1)

	// Connect ports
	require.NoError(t, ports.Connect(filter.InPort, inCh))
	require.NoError(t, ports.Connect(filter.OutPort, outCh))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start processing in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- filter.Process(ctx)
	}()

	// Test cases
	testCases := []struct {
		input    int
		expected bool // whether we expect the value to pass through
	}{
		{2, true},
		{3, false},
		{4, true},
		{5, false},
	}

	for _, tc := range testCases {
		// Send test data
		require.NoError(t, filter.InPort.Send(ctx, ip.New(tc.input)))

		if tc.expected {
			// Should receive output
			select {
			case packet := <-outCh:
				assert.Equal(t, tc.input, packet.Data())
			case <-time.After(time.Second):
				t.Fatalf("timeout waiting for output of %d", tc.input)
			}
		} else {
			// Should not receive output
			select {
			case packet := <-outCh:
				t.Fatalf("unexpected output received: %d", packet.Data())
			case <-time.After(100 * time.Millisecond):
				// This is expected
			}
		}
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

func TestFilterNilPredicate(t *testing.T) {
	filter := NewFilter[int](nil)
	err := filter.Process(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil predicate")
}
