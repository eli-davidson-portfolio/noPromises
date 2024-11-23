package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name      string
		requests  int
		interval  time.Duration
		limit     int
		wantBlock bool
	}{
		{
			name:      "under limit",
			requests:  5,
			interval:  time.Second,
			limit:     10,
			wantBlock: false,
		},
		{
			name:      "at limit",
			requests:  10,
			interval:  time.Second,
			limit:     10,
			wantBlock: false,
		},
		{
			name:      "over limit",
			requests:  15,
			interval:  time.Second,
			limit:     10,
			wantBlock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			limiter := NewRateLimiter(tt.limit, tt.interval)
			limitedHandler := limiter.Middleware(handler)

			var blocked bool
			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				w := httptest.NewRecorder()

				limitedHandler.ServeHTTP(w, req)

				if w.Code == http.StatusTooManyRequests {
					blocked = true
					break
				}
			}

			assert.Equal(t, tt.wantBlock, blocked)
		})
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	limiter := NewRateLimiter(50, time.Second)
	limitedHandler := limiter.Middleware(handler)

	// Run 100 concurrent requests
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			limitedHandler.ServeHTTP(w, req)
			done <- w.Code != http.StatusTooManyRequests
		}()
	}

	// Count successful requests
	successful := 0
	for i := 0; i < 100; i++ {
		if <-done {
			successful++
		}
	}

	// Should have exactly 50 successful requests
	assert.Equal(t, 50, successful)
}
