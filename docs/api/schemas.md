# API Schemas

## Authentication

### Token Request
```json
{
    "type": "object",
    "required": ["username", "password"],
    "properties": {
        "username": {
            "type": "string",
            "minLength": 1
        },
        "password": {
            "type": "string",
            "minLength": 1
        }
    }
}
```

### Token Response
```json
{
    "type": "object",
    "required": ["token"],
    "properties": {
        "token": {
            "type": "string",
            "description": "JWT token"
        }
    }
}
```

### Error Response
```json
{
    "type": "object",
    "required": ["error"],
    "properties": {
        "error": {
            "type": "string",
            "description": "Error message"
        }
    }
}
```

## API Error Types

```go
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
```