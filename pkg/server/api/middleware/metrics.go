package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// Metrics interface defines methods for recording metrics
type Metrics interface {
	RecordRequest(method, path string)
	RecordRequestDuration(duration time.Duration)
	RecordResponseStatus(status int)
	RecordFlowCreation(flowID string)
	RecordFlowDeletion(flowID string)
	RecordFlowStart(flowID string)
	RecordFlowStop(flowID string)
}

// metricsResponseWriter wraps http.ResponseWriter to capture the status code
type metricsResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *metricsResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// MetricsMiddleware creates middleware for recording request metrics
func MetricsMiddleware(m Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			m.RecordRequest(r.Method, r.URL.Path)

			// Create response wrapper to capture status code
			rw := &metricsResponseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			m.RecordRequestDuration(duration)
			m.RecordResponseStatus(rw.status)

			// Record labels after response is complete
			if m, ok := m.(*mockMetrics); ok {
				m.AddLabels(map[string]string{
					"method": r.Method,
					"path":   r.URL.Path,
					"status": fmt.Sprintf("%d", rw.status),
				})
			}
		})
	}
}
