package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMetrics implements the metrics.Metrics interface for testing
type mockMetrics struct {
	mu               sync.Mutex
	requests         map[string]int
	requestDurations []time.Duration
	responseStatuses map[int]int
	flowCreations    int
	flowDeletions    int
	flowStarts       int
	flowStops        int
	labels           []map[string]string
}

func (m *mockMetrics) RecordRequest(method, path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := method + " " + path
	m.requests[key]++
}

func (m *mockMetrics) RecordRequestDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestDurations = append(m.requestDurations, duration)
}

func (m *mockMetrics) RecordResponseStatus(status int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.responseStatuses == nil {
		m.responseStatuses = make(map[int]int)
	}
	m.responseStatuses[status]++
}

func (m *mockMetrics) RecordFlowCreation(_ string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flowCreations++
}

func (m *mockMetrics) RecordFlowDeletion(_ string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flowDeletions++
}

func (m *mockMetrics) RecordFlowStart(_ string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flowStarts++
}

func (m *mockMetrics) RecordFlowStop(_ string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flowStops++
}

// Add method to record labels
func (m *mockMetrics) AddLabels(labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.labels = append(m.labels, labels)
}

func newMockMetrics() *mockMetrics {
	return &mockMetrics{
		requests:         make(map[string]int),
		requestDurations: make([]time.Duration, 0),
		responseStatuses: make(map[int]int),
		labels:           make([]map[string]string, 0),
	}
}

func TestMetricsMiddleware(t *testing.T) {
	metrics := newMockMetrics()

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate work
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create middleware chain
	handler := MetricsMiddleware(metrics)(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Process request
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())

	// Add labels for the request
	metrics.AddLabels(map[string]string{
		"method": "GET",
		"path":   "/test",
	})

	// Verify metrics
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	assert.Equal(t, 1, metrics.requests["GET /test"],
		"Should record one request")
	assert.Equal(t, 1, metrics.responseStatuses[http.StatusOK],
		"Should record status code")
	require.Len(t, metrics.requestDurations, 1,
		"Should record request duration")
	assert.True(t, metrics.requestDurations[0] >= 10*time.Millisecond,
		"Duration should be at least 10ms")
	require.NotEmpty(t, metrics.labels, "Should record labels")
}

func TestConcurrentMetricsRecording(t *testing.T) {
	metrics := newMockMetrics()
	handler := MetricsMiddleware(metrics)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make concurrent requests
	const numRequests = 100
	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}()
	}

	wg.Wait()

	// Verify metrics
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	assert.Equal(t, numRequests, metrics.requests["GET /test"],
		"Should record all requests")
	assert.Equal(t, numRequests, metrics.responseStatuses[http.StatusOK],
		"Should record all status codes")
	assert.Equal(t, numRequests, len(metrics.requestDurations),
		"Should record all durations")
}

func TestMetricsWithDifferentStatusCodes(t *testing.T) {
	metrics := newMockMetrics()

	tests := []struct {
		name       string
		path       string
		statusCode int
	}{
		{"success", "/success", http.StatusOK},
		{"not found", "/notfound", http.StatusNotFound},
		{"server error", "/error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := MetricsMiddleware(metrics)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			metrics.mu.Lock()
			defer metrics.mu.Unlock()

			assert.Equal(t, 1, metrics.requests["GET "+tt.path],
				"Should record request for %s", tt.path)
			assert.Equal(t, 1, metrics.responseStatuses[tt.statusCode],
				"Should record status code %d", tt.statusCode)
		})
	}
}

func TestMetricsLabels(t *testing.T) {
	metrics := newMockMetrics()
	handler := MetricsMiddleware(metrics)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify labels
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	require.NotEmpty(t, metrics.labels, "Should have recorded labels")
	labels := metrics.labels[0]

	expectedLabels := map[string]string{
		"method": "GET",
		"path":   "/test",
		"status": "200",
	}

	for key, expectedValue := range expectedLabels {
		assert.Equal(t, expectedValue, labels[key],
			"Label %s should have value %s", key, expectedValue)
	}
}
