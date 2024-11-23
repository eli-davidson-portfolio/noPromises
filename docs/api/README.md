# noPromises API Documentation

## Core Components

### API Server
The API server provides a RESTful interface with the following key features:

- Authentication middleware with JWT support
- Rate limiting
- Error handling
- Database integration

### Key Interfaces

```go
// DB interface defines required database operations
type DB interface {
    // Add database methods as needed
}

// API represents the main API server
type API struct {
    db     DB
    router http.Handler
}
```

### Configuration

The API server can be configured using functional options:

```go
api := New(
    WithDB(db),
    // Add other options as needed
)
```

## Authentication

Authentication is handled via JWT tokens with the following features:

- Bearer token authentication
- Token generation endpoint
- Protected routes
- Claims-based authorization

### Authentication Flow

1. Client requests token via `/token` endpoint
2. Server validates credentials and issues JWT
3. Client includes token in `Authorization: Bearer <token>` header
4. Server validates token and processes request

### Rate Limiting

The API includes a token bucket rate limiter:
- Configurable rate and window
- Per-endpoint rate limits
- Automatic request throttling

## Error Handling

Standardized error responses:

```json
{
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable message"
    }
}
```

Common error types:
- Invalid credentials
- Missing/invalid token  
- Rate limit exceeded
- Internal server errors

## Testing

The API includes comprehensive test coverage:
- Unit tests for all components
- Concurrent operation testing
- Authentication flow testing
- Rate limiter testing
