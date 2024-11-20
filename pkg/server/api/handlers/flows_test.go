package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Add mock flow manager for testing
type mockFlowManager struct{}

func (m *mockFlowManager) CreateFlow(_ string, _ map[string]interface{}) error {
	return nil
}

func (m *mockFlowManager) GetFlow(id string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"id": id,
	}, nil
}

func TestFlowHandlers(t *testing.T) {
	t.Run("create flow", func(t *testing.T) {

		handler := NewFlowHandler(&mockFlowManager{})

		// Test request
		req := httptest.NewRequest(http.MethodPost, "/flows", strings.NewReader(`{
			"id": "test-flow",
			"nodes": {
				"reader": {
					"type": "FileReader",
					"config": {
						"filename": "test.txt"
					}
				}
			}
		}`))
		w := httptest.NewRecorder()

		handler.CreateFlow(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "test-flow", resp["id"])
	})
}
