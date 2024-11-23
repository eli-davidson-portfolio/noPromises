package api

// Error represents an API error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}

// Common API errors
var (
	ErrInvalidCredentials = &Error{
		Code:    "INVALID_CREDENTIALS",
		Message: "Invalid username or password",
	}
	ErrMissingToken = &Error{
		Code:    "MISSING_TOKEN",
		Message: "Missing authorization token",
	}
	ErrInvalidToken = &Error{
		Code:    "INVALID_TOKEN",
		Message: "Invalid token",
	}
	ErrRateLimitExceeded = &Error{
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: "Rate limit exceeded",
	}
)
