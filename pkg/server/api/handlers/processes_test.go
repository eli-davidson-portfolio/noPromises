package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListProcesses(t *testing.T) {
	handler := &ProcessHandler{}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/processes", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ListProcesses(w, req)

	// Verify response
	require.Equal(t, http.StatusNotImplemented, w.Code)

	var resp map[string]string
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Contains(t, resp, "error")
}
