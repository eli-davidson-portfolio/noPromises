package process

import (
	"context"
	"testing"
	"time"
)

func TestBaseProcess(t *testing.T) {
	t.Run("name", func(t *testing.T) {
		name := "test-process"
		proc := NewBaseProcess(name)
		if proc.Name() != name {
			t.Errorf("Expected name %q, got %q", name, proc.Name())
		}
	})

	t.Run("process", func(t *testing.T) {
		proc := NewBaseProcess("test")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- proc.Process(ctx)
		}()

		select {
		case err := <-errCh:
			if err != context.DeadlineExceeded {
				t.Errorf("Expected DeadlineExceeded, got %v", err)
			}
		case <-time.After(200 * time.Millisecond):
			t.Error("Process didn't return after context cancellation")
		}
	})
}
