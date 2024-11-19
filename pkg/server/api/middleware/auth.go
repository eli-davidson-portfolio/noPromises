package middleware

import (
	"net/http"
)

// AuthMiddleware provides authentication for API endpoints
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement authentication
		next.ServeHTTP(w, r)
	})
}
