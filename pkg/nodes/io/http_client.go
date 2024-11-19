package io

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/nodes"
)

// HTTPClient makes HTTP requests and forwards the responses
type HTTPClient struct {
	*nodes.BaseNode[string, []byte]
	client *http.Client
}

// NewHTTPClient creates a new HTTP client node
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		BaseNode: nodes.NewBaseNode[string, []byte]("HTTPClient"),
		client:   &http.Client{},
	}
}

// Process implements the processing logic
func (h *HTTPClient) Process(ctx context.Context) error {
	if h.client == nil {
		return fmt.Errorf("nil HTTP client")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			packet, err := h.InPort.Receive(ctx)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					return err
				}
				return fmt.Errorf("receive failed: %w", err)
			}

			req, err := http.NewRequestWithContext(ctx, "GET", packet.Data(), nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			resp, err := h.client.Do(req)
			if err != nil {
				// Check if the error is due to context cancellation
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := h.OutPort.Send(ctx, ip.New(body)); err != nil {
					if err == context.Canceled || err == context.DeadlineExceeded {
						return err
					}
					return fmt.Errorf("send failed: %w", err)
				}
			}
		}
	}
}
