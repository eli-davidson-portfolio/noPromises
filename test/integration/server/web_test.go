package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebIntegration(t *testing.T) {
	srv := setupTestServer(t)

	t.Run("serve home page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Test Documentation")
	})

	t.Run("serve docs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs/README.md", nil)
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Test Documentation")
	})

	t.Run("serve swagger", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/swagger.json", nil)
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})
}
