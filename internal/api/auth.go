// internal/api/auth.go
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	logger *zap.Logger
	secret []byte
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger *zap.Logger, secret []byte) *AuthHandler {
	return &AuthHandler{
		logger: logger,
		secret: secret,
	}
}

// getBearerToken extracts the token from the Authorization header
func getBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

// validateToken validates the JWT token
func validateToken(tokenString string, secret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// respondError sends an error response with the given status code and message
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	}); err != nil {
		log.Printf("failed to encode error response: %v", err)
	}
}

// AuthMiddleware provides authentication for protected endpoints
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for certain endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/token" {
			next.ServeHTTP(w, r)
			return
		}

		token := getBearerToken(r)
		if token == "" {
			respondError(w, http.StatusUnauthorized, "missing authorization token")
			return
		}

		claims, err := validateToken(token, h.secret)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		// Add claims to request context
		ctx := setUserContext(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HandleToken handles token generation requests
func (h *AuthHandler) HandleToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	// For testing purposes, accept any non-empty credentials
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(h.secret)
	if err != nil {
		h.logger.Error("failed to sign token", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	}); err != nil {
		h.logger.Error("failed to encode token response", zap.Error(err))
		return
	}
}
