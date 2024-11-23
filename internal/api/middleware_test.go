package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuthMiddleware(t *testing.T) {
	logger := zap.NewNop()
	secret := []byte("test-secret")
	handler := NewAuthHandler(logger, secret)

	tests := []struct {
		name       string
		token      string
		path       string
		wantStatus int
	}{
		{
			name:       "valid token",
			token:      generateTestToken(t, secret),
			path:       "/protected",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing token",
			token:      "",
			path:       "/protected",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			token:      "invalid.token.here",
			path:       "/protected",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "health check bypass",
			token:      "",
			path:       "/health",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Apply middleware
			middlewareHandler := handler.AuthMiddleware(testHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()

			// Handle request
			middlewareHandler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// Helper function to generate a test token
func generateTestToken(t *testing.T, secret []byte) string {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "test",
	}).SignedString(secret)
	require.NoError(t, err)
	return token
}
