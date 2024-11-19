package io

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/core/ports"
	"github.com/elleshadow/noPromises/pkg/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPClient(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("test response"))
	}))
	defer ts.Close()

	client := NewHTTPClient()

	// Create test channels
	inCh := make(chan *ip.IP[string], 1)
	outCh := make(chan *ip.IP[[]byte], 1)

	// Connect ports
	require.NoError(t, ports.Connect(client.InPort, inCh))
	require.NoError(t, ports.Connect(client.OutPort, outCh))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start processing in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Process(ctx)
	}()

	// Send test URL
	require.NoError(t, client.InPort.Send(ctx, ip.New(ts.URL)))

	// Verify response
	select {
	case packet := <-outCh:
		assert.Equal(t, []byte("test response"), packet.Data())
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for response")
	}

	// Test invalid URL
	require.NoError(t, client.InPort.Send(ctx, ip.New("invalid-url")))
	select {
	case err := <-errCh:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request failed")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for error")
	}
}

func TestHTTPClientNilClient(t *testing.T) {
	client := &HTTPClient{
		BaseNode: nodes.NewBaseNode[string, []byte]("HTTPClient"),
	}
	err := client.Process(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil HTTP client")
}

func TestHTTPClientCancellation(t *testing.T) {
	// Create a server that delays response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Write([]byte("delayed response"))
	}))
	defer ts.Close()

	client := NewHTTPClient()

	// Create test channels
	inCh := make(chan *ip.IP[string], 1)
	outCh := make(chan *ip.IP[[]byte], 1)

	// Connect ports
	require.NoError(t, ports.Connect(client.InPort, inCh))
	require.NoError(t, ports.Connect(client.OutPort, outCh))

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start processing in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Process(ctx)
	}()

	// Send test URL
	require.NoError(t, client.InPort.Send(ctx, ip.New(ts.URL)))

	// Verify context cancellation
	select {
	case err := <-errCh:
		assert.Equal(t, context.DeadlineExceeded, err)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for cancellation")
	}
}
