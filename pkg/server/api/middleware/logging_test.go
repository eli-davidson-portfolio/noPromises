package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddleware(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create middleware chain
	handler := LoggingMiddleware(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Process request
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())

	// Verify log output
	logOutput := logBuf.String()
	require.True(t, strings.Contains(logOutput, "GET /test 200"),
		"Log should contain request method, path and status code")
}

func TestResponseWriterWrapper(t *testing.T) {
	t.Run("explicit_status_code", func(t *testing.T) {
		rw := wrapResponseWriter(httptest.NewRecorder())
		rw.WriteHeader(http.StatusNotFound)
		assert.Equal(t, http.StatusNotFound, rw.status)
	})

	t.Run("default_status_code", func(t *testing.T) {
		rw := wrapResponseWriter(httptest.NewRecorder())
		rw.Write([]byte("test"))
		assert.Equal(t, http.StatusOK, rw.status)
	})

	t.Run("multiple_writes", func(t *testing.T) {
		rw := wrapResponseWriter(httptest.NewRecorder())
		rw.WriteHeader(http.StatusBadRequest)
		rw.WriteHeader(http.StatusOK)
		assert.Equal(t, http.StatusBadRequest, rw.status)
	})
}

func TestConcurrentLogging(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)

	// Create test handler with delay
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware chain
	handler := LoggingMiddleware(testHandler)

	// Make concurrent requests
	const numRequests = 10
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(_ int) {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}

	// Verify log output contains all requests
	logOutput := logBuf.String()
	logLines := strings.Split(strings.TrimSpace(logOutput), "\n")
	assert.Equal(t, numRequests, len(logLines),
		"Should have logged exactly %d requests", numRequests)
}
