package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type tokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestAuthHandler(t *testing.T) {
	logger := zap.NewNop()
	secret := []byte("test-secret")
	handler := NewAuthHandler(logger, secret)

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
		wantToken  bool
	}{
		{
			name:       "valid credentials",
			username:   "test",
			password:   "password",
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name:       "missing username",
			password:   "password",
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
		{
			name:       "missing password",
			username:   "test",
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := tokenRequest{
				Username: tt.username,
				Password: tt.password,
			}
			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.HandleToken(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var resp map[string]string
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			if tt.wantToken {
				assert.NotEmpty(t, resp["token"])
			} else {
				assert.Empty(t, resp["token"])
			}
		})
	}
}
