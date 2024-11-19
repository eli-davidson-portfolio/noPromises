package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.Handler
		wantStatus     int
		wantPanic      bool
		wantErrMessage string
	}{
		{
			name: "normal request",
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}),
			wantStatus: http.StatusOK,
			wantPanic:  false,
		},
		{
			name: "panic with error",
			handler: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				panic("test panic")
			}),
			wantStatus:     http.StatusInternalServerError,
			wantPanic:      true,
			wantErrMessage: "Internal Server Error",
		},
		{
			name: "panic with custom error",
			handler: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				panic(struct{ msg string }{msg: "custom error"})
			}),
			wantStatus:     http.StatusInternalServerError,
			wantPanic:      true,
			wantErrMessage: "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create recovery middleware
			handler := RecoveryMiddleware(tt.handler)

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Process request
			handler.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantPanic {
				var resp map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)

				errObj, ok := resp["error"].(map[string]interface{})
				require.True(t, ok, "Response should contain error object")
				assert.Equal(t, tt.wantErrMessage, errObj["message"])
			} else {
				if w.Code == http.StatusOK {
					assert.Equal(t, "success", w.Body.String())
				}
			}
		})
	}
}

func TestRecoveryConcurrent(t *testing.T) {
	// Create handler that randomly panics
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("panic") == "true" {
			panic("random panic")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with recovery middleware
	recoveryHandler := RecoveryMiddleware(handler)

	// Make concurrent requests
	const numRequests = 100
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(shouldPanic bool) {
			req := httptest.NewRequest("GET", "/test?panic="+fmt.Sprint(shouldPanic), nil)
			w := httptest.NewRecorder()
			recoveryHandler.ServeHTTP(w, req)
			results <- w.Code
		}(i%2 == 0) // Alternate between panic and success
	}

	// Collect results
	okCount := 0
	errorCount := 0
	for i := 0; i < numRequests; i++ {
		status := <-results
		switch status {
		case http.StatusOK:
			okCount++
		case http.StatusInternalServerError:
			errorCount++
		default:
			t.Errorf("Unexpected status code: %d", status)
		}
	}

	// Verify results
	assert.Equal(t, numRequests/2, okCount, "Should have expected number of successful requests")
	assert.Equal(t, numRequests/2, errorCount, "Should have expected number of recovered panics")
}

func TestRecoveryWithNestedPanics(t *testing.T) {
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("initial panic")
	})

	recoveryHandler := RecoveryMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	recoveryHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	errObj, ok := resp["error"].(map[string]interface{})
	require.True(t, ok, "Response should contain error object")
	assert.Equal(t, "Internal Server Error", errObj["message"])
}
