# API Design Patterns

## Overview

This document outlines the key patterns and principles used in designing the noPromises API. These patterns ensure consistency, maintainability, and a good developer experience.

## Core Principles

### 1. Resource-Oriented Design

All API endpoints are organized around resources:

```
/api/v1/flows           # Collection of flows
/api/v1/flows/{id}      # Individual flow
/api/v1/processes       # Collection of process types
/api/v1/processes/{id}  # Individual process type
```

Key principles:
- Resources are nouns, not verbs
- Use plural forms for collections
- Nest related resources when it makes sense
- Keep URLs intuitive and predictable

### 2. Standard HTTP Methods

Use HTTP methods according to their standard semantics:

| Method | Usage | Example |
|--------|--------|---------|
| GET    | Retrieve resource(s) | `GET /api/v1/flows` |
| POST   | Create new resource | `POST /api/v1/flows` |
| PUT    | Replace resource | `PUT /api/v1/flows/{id}` |
| DELETE | Remove resource | `DELETE /api/v1/flows/{id}` |

Special cases:
- Use POST for actions that don't fit CRUD: `/api/v1/flows/{id}/start`
- Avoid overloading GET with complex query parameters
- Use PATCH for partial updates (future enhancement)

### 3. Consistent Response Structure

All responses follow a consistent structure:

```json
// Success with data
{
    "data": {
        // Resource-specific data
    },
    "metadata": {
        // Optional metadata
    }
}

// Error response
{
    "error": {
        "code": "ERROR_CODE",
        "message": "Human-readable message",
        "details": {
            // Additional error details
        }
    }
}
```

### 4. HTTP Status Codes

Use appropriate HTTP status codes:

| Code | Usage |
|------|--------|
| 200  | Success (GET, PUT) |
| 201  | Resource created (POST) |
| 204  | Success, no content (DELETE) |
| 400  | Invalid request |
| 401  | Unauthorized |
| 403  | Forbidden |
| 404  | Resource not found |
| 409  | Conflict |
| 422  | Validation error |
| 429  | Rate limit exceeded |
| 500  | Server error |

### 5. Query Parameters

Standard query parameters:

```
?limit=20           # Limit results
?offset=40         # Pagination offset
?sort=created_at   # Sort field
?order=desc        # Sort order
?fields=id,name    # Field selection
```

Collection filtering:
```
?state=running     # Filter by state
?created_after=... # Time-based filtering
?type=FileReader   # Type filtering
```

### 6. Versioning

Version through URL path:
```
/api/v1/...  # Current version
/api/v2/...  # Future version
```

Include version in response headers:
```
X-API-Version: 1.0
```

### 7. Rate Limiting

Include rate limit headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1573664000
```

### 8. Request/Response Headers

Standard request headers:
```
Content-Type: application/json
Accept: application/json
Authorization: Bearer <token>
X-Request-ID: <client-generated-id>
```

Standard response headers:
```
Content-Type: application/json
X-Request-ID: <echoed-from-request>
X-API-Version: 1.0
```

## Implementation Patterns

### 1. Request Validation

Multi-stage validation:

```go
type Handler struct {
    validator *Validator
    engine    FlowEngine
}

func (h *Handler) CreateFlow(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    var req CreateFlowRequest
    if err := parseJSON(r, &req); err != nil {
        respondError(w, NewParseError(err))
        return
    }

    // 2. Validate schema
    if err := h.validator.ValidateFlowConfig(req); err != nil {
        respondError(w, NewValidationError(err))
        return
    }

    // 3. Business logic validation
    if err := h.validateFlowLogic(req); err != nil {
        respondError(w, NewBusinessError(err))
        return
    }

    // 4. Process request
    flow, err := h.engine.CreateFlow(req.ToConfig())
    if err != nil {
        respondError(w, err)
        return
    }

    // 5. Send response
    respondJSON(w, http.StatusCreated, flow)
}
```

### 2. Error Handling

Structured error types:

```go
type APIError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func NewValidationError(err error) *APIError {
    return &APIError{
        Code:    "VALIDATION_ERROR",
        Message: "Request validation failed",
        Details: map[string]interface{}{
            "errors": err.Error(),
        },
    }
}
```

### 3. Response Helpers

Consistent response formatting:

```go
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "data": data,
    })
}

func respondError(w http.ResponseWriter, err error) {
    apiErr, status := toAPIError(err)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": apiErr,
    })
}
```

### 4. Middleware Chain

Standard middleware stack:

```go
router.Use(
    middleware.RequestID,          // Add request ID
    middleware.RealIP,            // Extract real IP
    middleware.Logger,            // Log requests
    middleware.Recoverer,         // Recover from panics
    middleware.Timeout(60*time.Second),
    rateLimiter.Handler,          // Rate limiting
    auth.Handler,                 // Authentication
)
```

### 5. Request Context

Use context for request-scoped data:

```go
type contextKey string

const (
    userContextKey    = contextKey("user")
    requestIDKey      = contextKey("request-id")
    correlationIDKey  = contextKey("correlation-id")
)

func (h *Handler) GetFlow(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value(userContextKey).(string)
    requestID := r.Context().Value(requestIDKey).(string)
    
    // Use context values in processing
}
```

## Testing Patterns

### 1. Table-Driven Tests

```go
func TestCreateFlow(t *testing.T) {
    tests := []struct {
        name    string
        req     CreateFlowRequest
        want    *Flow
        wantErr bool
    }{
        {
            name: "valid flow",
            req:  CreateFlowRequest{...},
            want: &Flow{...},
        },
        {
            name:    "invalid config",
            req:     CreateFlowRequest{...},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := handler.CreateFlow(tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateFlow() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("CreateFlow() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 2. Mock Server

```go
type MockServer struct {
    Engine FlowEngine
    mux    *http.ServeMux
}

func NewMockServer(engine FlowEngine) *MockServer {
    s := &MockServer{
        Engine: engine,
        mux:    http.NewServeMux(),
    }
    s.routes()
    return s
}

func (s *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.mux.ServeHTTP(w, r)
}
```

## Security Patterns

### 1. Authentication

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractToken(r)
        if token == "" {
            respondError(w, ErrUnauthorized)
            return
        }

        user, err := validateToken(token)
        if err != nil {
            respondError(w, ErrUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), userContextKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 2. Input Sanitization

```go
func sanitizeInput(input string) string {
    // Remove potentially dangerous characters
    return strings.Map(func(r rune) rune {
        if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '_' {
            return r
        }
        return -1
    }, input)
}
```

## Future Considerations

1. **API Evolution**
   - Maintain backward compatibility
   - Use feature flags for new functionality
   - Plan for versioning transitions

2. **Performance Optimizations**
   - Response caching
   - Batch operations
   - Streaming responses

3. **Advanced Features**
   - WebSocket support for real-time updates
   - GraphQL API for complex queries
   - Bulk operations support