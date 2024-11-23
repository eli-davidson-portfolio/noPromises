package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError(t *testing.T) {
	tests := []struct {
		name    string
		err     *Error
		wantMsg string
	}{
		{
			name: "basic error",
			err: &Error{
				Code:    "TEST_ERROR",
				Message: "test error message",
			},
			wantMsg: "test error message",
		},
		{
			name:    "auth error",
			err:     ErrInvalidCredentials,
			wantMsg: "Invalid username or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMsg, tt.err.Error())
		})
	}
}
