# Server API Subsystem

## Overview

The Server API subsystem provides the HTTP interface to the noPromises Flow-Based Programming (FBP) system. It translates HTTP requests into FlowEngine operations and manages all aspects of the RESTful API.

## Core Components

### Server Structure

```go
// Server is the main API server component
type Server struct {
    engine    FlowEngine
    router    chi.Router
    validator *Validator
    logger    *Logger
    metrics   *Metrics
    config    *Config
}

// Config holds server configuration
type Config struct {
    Address     string
    BasePath    string
    Timeout     time.Duration
    RateLimit   RateLimitConfig
    Cors        CorsConfig
    Auth        AuthConfig
}

// NewServer creates a new API server
func NewServer(engine FlowEngine, config *Config) *Server {
    s := &Server{
        engine:    engine,
        router:    chi.NewRouter(),
        validator: NewValidator(),
        logger:    NewLogger(),
        metrics:   NewMetrics(),
        config:    config,
    }
    s.setupRoutes()
    s.setupMiddleware()
    return s
}
```

### Route Registration

```go
func (s *Server) setupRoutes() {
    // API version group
    s.router.Route("/api/v1", func(r chi.Router) {
        // Flow management
        r.Route("/flows", func(r chi.Router) {
            r.Get("/", s.ListFlows)
            r.Post("/", s.CreateFlow)
            r.Route("/{flowID}", func(r chi.Router) {
                r.Use(s.FlowCtx)  // Add flow to context
                r.Get("/", s.GetFlow)
                r.Delete("/", s.DeleteFlow)
                r.Post("/start", s.StartFlow)
                r.Post("/stop", s.StopFlow)
                r.Get("/status", s.GetFlowStatus)
            })
        })

        // Process type discovery
        r.Route("/processes", func(r chi.Router) {
            r.Get("/", s.ListProcesses)
            r.Get("/{processType}", s.GetProcessDetails)
        })
    })
}
```

### Middleware Setup

```go
func (s *Server) setupMiddleware() {
    s.router.Use(
        middleware.RequestID,
        middleware.RealIP,
        s.loggerMiddleware,
        middleware.Recoverer,
        middleware.Timeout(s.config.Timeout),
        s.corsMiddleware,
        s.rateLimitMiddleware,
        s.authMiddleware,
        middleware.AllowContentType("application/json"),
    )
}
```

## Request Handling

### Context Management

```go
// FlowCtx middleware adds flow ID to request context
func (s *Server) FlowCtx(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        flowID := chi.URLParam(r, "flowID")
        ctx := context.WithValue(r.Context(), flowIDKey, flowID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetFlowFromContext retrieves flow ID from context
func GetFlowFromContext(ctx context.Context) string {
    if flowID, ok := ctx.Value(flowIDKey).(string); ok {
        return flowID
    }
    return ""
}
```

### Request Validation

```go
// validateRequest validates request body against JSON schema
func (s *Server) validateRequest(r *http.Request, schema string) error {
    var data interface{}
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        return NewParseError(err)
    }
    
    if err := s.validator.Validate(schema, data); err != nil {
        return NewValidationError(err)
    }
    
    return nil
}
```

### Response Handling

```go
// Response wraps all successful responses
type Response struct {
    Data     interface{}            `json:"data,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// respondJSON sends a JSON response
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    response := Response{
        Data: data,
        Metadata: map[string]interface{}{
            "timestamp": time.Now().UTC(),
            "version":   "1.0",
        },
    }
    
    if err := json.NewEncoder(w).Encode(response); err != nil {
        s.logger.Error("failed to encode response", "error", err)
    }
}

// respondError sends an error response
func (s *Server) respondError(w http.ResponseWriter, err error) {
    apiErr, status := toAPIError(err)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    if err := json.NewEncoder(w).Encode(apiErr); err != nil {
        s.logger.Error("failed to encode error response", "error", err)
    }
}
```

## Handler Implementations

### Flow Management

```go
// CreateFlow handles flow creation requests
func (s *Server) CreateFlow(w http.ResponseWriter, r *http.Request) {
    // Validate request
    var req CreateFlowRequest
    if err := s.validateRequest(r, "flow-config"); err != nil {
        s.respondError(w, err)
        return
    }

    // Create flow
    flow, err := s.engine.CreateFlow(req.ToConfig())
    if err != nil {
        s.respondError(w, err)
        return
    }

    // Send response
    s.respondJSON(w, http.StatusCreated, flow)
}

// StartFlow handles flow start requests
func (s *Server) StartFlow(w http.ResponseWriter, r *http.Request) {
    flowID := GetFlowFromContext(r.Context())
    
    if err := s.engine.StartFlow(flowID); err != nil {
        s.respondError(w, err)
        return
    }
    
    s.respondJSON(w, http.StatusOK, map[string]string{
        "status": "started",
    })
}
```

### Process Discovery

```go
// ListProcesses handles process type listing
func (s *Server) ListProcesses(w http.ResponseWriter, r *http.Request) {
    processes, err := s.engine.ListAvailableProcessTypes()
    if err != nil {
        s.respondError(w, err)
        return
    }
    
    s.respondJSON(w, http.StatusOK, processes)
}
```

## Error Handling

```go
// APIError represents an API error response
type APIError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// toAPIError converts an error to an APIError and status code
func toAPIError(err error) (*APIError, int) {
    switch e := err.(type) {
    case *ValidationError:
        return &APIError{
            Code:    "VALIDATION_ERROR",
            Message: "Request validation failed",
            Details: map[string]interface{}{
                "errors": e.Errors(),
            },
        }, http.StatusBadRequest
        
    case *NotFoundError:
        return &APIError{
            Code:    "NOT_FOUND",
            Message: e.Error(),
        }, http.StatusNotFound
        
    // Add other error types...
        
    default:
        return &APIError{
            Code:    "INTERNAL_ERROR",
            Message: "An internal error occurred",
        }, http.StatusInternalServerError
    }
}
```

## Metrics and Monitoring

```go
// MetricsMiddleware adds request metrics
func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap response writer to capture status code
        ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
        
        next.ServeHTTP(ww, r)
        
        // Record metrics
        s.metrics.RecordRequest(r.Method, r.URL.Path, ww.Status(), time.Since(start))
    })
}
```

## Testing Support

```go
// TestServer provides a test server
type TestServer struct {
    *Server
    Engine *MockEngine
}

// NewTestServer creates a new test server
func NewTestServer(t *testing.T) *TestServer {
    engine := NewMockEngine(t)
    server := NewServer(engine, &Config{
        Address:  "localhost:0",
        BasePath: "/api/v1",
    })
    
    return &TestServer{
        Server: server,
        Engine: engine,
    }
}

// MockEngine implements FlowEngine for testing
type MockEngine struct {
    t *testing.T
    mock.Mock
}
```

## Usage Example

```go
func main() {
    // Create flow engine
    engine := fbp.NewFlowEngine()

    // Create server
    server := NewServer(engine, &Config{
        Address:  ":8080",
        BasePath: "/api/v1",
        Timeout:  60 * time.Second,
        RateLimit: RateLimitConfig{
            RequestsPerMinute: 100,
        },
    })

    // Start server
    log.Printf("Starting server on %s", server.config.Address)
    if err := server.ListenAndServe(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

## Best Practices

1. **Request Validation**
   - Validate all request inputs
   - Use JSON schema validation
   - Sanitize inputs
   - Validate business logic

2. **Error Handling**
   - Use structured error types
   - Include appropriate details
   - Log internal errors
   - Return user-friendly messages

3. **Performance**
   - Use connection pooling
   - Enable response compression
   - Implement request timeouts
   - Monitor response times

4. **Security**
   - Validate content types
   - Implement rate limiting
   - Use secure headers
   - Sanitize outputs