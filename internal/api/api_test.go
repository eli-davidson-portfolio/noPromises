package api

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockDB implements the DB interface for testing
type mockDB struct{}

func TestAPI(t *testing.T) {
	t.Run("new api", func(t *testing.T) {
		api := New()
		assert.NotNil(t, api)
	})

	t.Run("with options", func(t *testing.T) {
		db := &mockDB{}
		api := New(WithDB(db))
		assert.NotNil(t, api)
		assert.NotNil(t, api.db)
	})

	t.Run("serve http", func(t *testing.T) {
		api := New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotImplemented, w.Code)
	})
}

func TestAPIConcurrent(t *testing.T) {
	api := New()
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotImplemented, w.Code)
		}()
	}
	wg.Wait()
}

func TestAPIOptions(t *testing.T) {
	t.Run("with database", func(t *testing.T) {
		db := &mockDB{}
		api := New(WithDB(db))
		assert.NotNil(t, api.db)
	})

	t.Run("without database", func(t *testing.T) {
		api := New()
		assert.Nil(t, api.db)
	})
}
