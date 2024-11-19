package process

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseProcessBasics(t *testing.T) {
	t.Run("name", func(t *testing.T) {
		p := NewBaseProcess("test")
		assert.Equal(t, "test", p.Name())
	})

	t.Run("process", func(t *testing.T) {
		p := NewBaseProcess("test")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := p.Process(ctx)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestProcessLifecycle(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		p := NewBaseProcess("test")
		ctx := context.Background()

		assert.False(t, p.IsInitialized())
		require.NoError(t, p.Initialize(ctx))
		assert.True(t, p.IsInitialized())
	})

	t.Run("shutdown", func(t *testing.T) {
		p := NewBaseProcess("test")
		ctx := context.Background()

		require.NoError(t, p.Initialize(ctx))
		assert.True(t, p.IsInitialized())

		require.NoError(t, p.Shutdown(ctx))
		assert.False(t, p.IsInitialized())

		err := p.Initialize(ctx)
		assert.Equal(t, ErrProcessShutdown, err)
	})
}

func TestProcessing(t *testing.T) {
	t.Run("basic transformation", func(t *testing.T) {
		p := NewBaseProcess("test")
		ctx := context.Background()

		require.NoError(t, p.Initialize(ctx))

		ctx, cancel := context.WithCancel(ctx)
		errCh := make(chan error, 1)
		go func() {
			errCh <- p.Process(ctx)
		}()

		cancel()
		assert.Equal(t, context.Canceled, <-errCh)
	})

	t.Run("context cancellation", func(t *testing.T) {
		p := NewBaseProcess("test")
		ctx := context.Background()

		require.NoError(t, p.Initialize(ctx))

		ctx, cancel := context.WithCancel(ctx)
		errCh := make(chan error, 1)
		go func() {
			errCh <- p.Process(ctx)
		}()

		cancel()
		assert.Equal(t, context.Canceled, <-errCh)
	})
}
