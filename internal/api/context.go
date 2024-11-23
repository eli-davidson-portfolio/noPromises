package api

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userClaimsKey contextKey = "userClaims"

// setUserContext adds JWT claims to the context
func setUserContext(ctx context.Context, claims jwt.MapClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}

// GetUserClaims retrieves JWT claims from context
func GetUserClaims(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(jwt.MapClaims)
	return claims, ok
}
