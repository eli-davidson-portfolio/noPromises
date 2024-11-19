package transform

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapper(t *testing.T) {
	// Create mapper that converts strings to uppercase
	toUpper := strings.ToUpper
	mapper := NewMapper[string, string](toUpper)

	// Create test channels
	inCh := make(chan *ip.IP[string], 1)
	outCh := make(chan *ip.IP[string], 1)

	// Connect ports
	require.NoError(t, ports.Connect(mapper.InPort, inCh))
	require.NoError(t, ports.Connect(mapper.OutPort, outCh))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start processing in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- mapper.Process(ctx)
	}()

	// Test cases
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello", "HELLO"},
		{"world", "WORLD"},
		{"Test", "TEST"},
	}

	for _, tc := range testCases {
		// Send test data
		require.NoError(t, mapper.InPort.Send(ctx, ip.New(tc.input)))

		// Verify output
		select {
		case packet := <-outCh:
			assert.Equal(t, tc.expected, packet.Data())
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for output of %q", tc.input)
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

func TestMapperNilTransform(t *testing.T) {
	mapper := NewMapper[string, string](nil)
	err := mapper.Process(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil transform")
}
