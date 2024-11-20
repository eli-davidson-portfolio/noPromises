package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowIntegration(t *testing.T) {
	srv := setupTestServer(t)

	// Test flow creation
	t.Run("create flow", func(t *testing.T) {
		flowConfig := map[string]interface{}{
			"id": "test-flow",
			"nodes": map[string]interface{}{
				"reader": map[string]interface{}{
					"type": "FileReader",
					"config": map[string]interface{}{
						"filename": "test.txt",
					},
				},
			},
		}

		err := srv.CreateFlow("test-flow", flowConfig)
		require.NoError(t, err)

		// Verify flow was created
		flow, exists := srv.GetFlow("test-flow")
		require.True(t, exists)
		assert.Equal(t, "test-flow", flow.ID)
	})

	// Test flow lifecycle
	t.Run("flow lifecycle", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := srv.StartFlow(ctx, "test-flow")
		require.NoError(t, err)

		// Verify flow is running
		flow, exists := srv.GetFlow("test-flow")
		require.True(t, exists)
		assert.Equal(t, "running", flow.State)

		err = srv.StopFlow(ctx, "test-flow")
		require.NoError(t, err)

		// Verify flow is stopped
		flow, exists = srv.GetFlow("test-flow")
		require.True(t, exists)
		assert.Equal(t, "stopped", flow.State)
	})
}
