package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs request details
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := wrapResponseWriter(w)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request details
		log.Printf(
			"%s %s %d %s",
			r.Method,
			r.RequestURI,
			wrapped.status,
			time.Since(start),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *responseWriter) WriteHeader(status int) {
	if !w.written {
		w.status = status
		w.written = true
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.status = http.StatusOK // Set default status if WriteHeader wasn't called
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}
