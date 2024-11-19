package nodes

import (
	"context"
	"testing"
	"time"
)

func TestBaseNode(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		node := NewBaseNode[string, int]("TestNode")

		if node == nil {
			t.Fatal("Expected non-nil node")
		}

		if node.InPort == nil {
			t.Error("Expected non-nil input port")
		}

		if node.OutPort == nil {
			t.Error("Expected non-nil output port")
		}

		if node.Config == nil {
			t.Error("Expected non-nil config map")
		}
	})

	t.Run("name", func(t *testing.T) {
		name := "TestNode"
		node := NewBaseNode[string, int](name)

		if node.Name() != name {
			t.Errorf("Expected name %q, got %q", name, node.Name())
		}
	})

	t.Run("process", func(t *testing.T) {
		node := NewBaseNode[string, int]("TestNode")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- node.Process(ctx)
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

	t.Run("config", func(t *testing.T) {
		node := NewBaseNode[string, int]("TestNode")

		// Test config operations
		node.Config["key"] = "value"
		if val, ok := node.Config["key"]; !ok || val != "value" {
			t.Errorf("Expected config value 'value', got %v", val)
		}
	})
}
