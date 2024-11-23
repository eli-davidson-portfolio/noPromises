package api

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	t.Run("set and get claims", func(t *testing.T) {
		claims := jwt.MapClaims{
			"username": "test-user",
			"role":     "admin",
		}

		ctx := context.Background()
		ctx = setUserContext(ctx, claims)

		gotClaims, ok := GetUserClaims(ctx)
		assert.True(t, ok, "should get claims successfully")
		assert.Equal(t, claims["username"], gotClaims["username"])
		assert.Equal(t, claims["role"], gotClaims["role"])
	})

	t.Run("get claims from empty context", func(t *testing.T) {
		ctx := context.Background()
		claims, ok := GetUserClaims(ctx)
		assert.False(t, ok, "should not get claims from empty context")
		assert.Nil(t, claims)
	})

	t.Run("context isolation", func(t *testing.T) {
		claims1 := jwt.MapClaims{"username": "user1"}
		claims2 := jwt.MapClaims{"username": "user2"}

		ctx1 := setUserContext(context.Background(), claims1)
		ctx2 := setUserContext(context.Background(), claims2)

		got1, ok := GetUserClaims(ctx1)
		assert.True(t, ok)
		assert.Equal(t, "user1", got1["username"])

		got2, ok := GetUserClaims(ctx2)
		assert.True(t, ok)
		assert.Equal(t, "user2", got2["username"])
	})
}
