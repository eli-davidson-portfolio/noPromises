package api

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu      sync.Mutex
	rate    int
	window  time.Duration
	tokens  int
	lastAdd time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rate:    rate,
		window:  window,
		tokens:  rate,
		lastAdd: time.Now(),
	}
}

// Middleware returns a rate limiting middleware
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow() {
			respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// allow checks if a request should be allowed
func (rl *RateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastAdd)

	// Reset if window has passed
	if elapsed >= rl.window {
		rl.tokens = rl.rate
		rl.lastAdd = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}
