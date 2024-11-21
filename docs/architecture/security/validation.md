# Input/Output Validation Architecture

## Core Components

### Validator Interface
```go
type Validator interface {
    // Validate validates data against schema
    Validate(schema string, data interface{}) error
    
    // ValidateJSON validates raw JSON data
    ValidateJSON(schema string, jsonData []byte) error
    
    // AddSchema adds or updates a named schema
    AddSchema(name string, schema string) error
}

// Implementation using Go validator
type JSONValidator struct {
    schemas    map[string]*jsonschema.Schema
    compiler   *jsonschema.Compiler
    mu         sync.RWMutex
}
```

### Schema Management
```go
// Process configuration schema
type ProcessSchema struct {
    Properties map[string]PropertySchema `json:"properties"`
    Required   []string                 `json:"required"`
}

type PropertySchema struct {
    Type        string      `json:"type"`
    Description string      `json:"description"`
    Default     any         `json:"default,omitempty"`
    Minimum     *float64    `json:"minimum,omitempty"`
    Maximum     *float64    `json:"maximum,omitempty"`
    Pattern     string      `json:"pattern,omitempty"`
}
```

## Validation Patterns

### Input Validation
```go
// Request validation middleware
func ValidateRequest(schema string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            var data interface{}
            if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
                respondError(w, NewValidationError("invalid JSON"))
                return
            }
            
            if err := validator.Validate(schema, data); err != nil {
                respondError(w, NewValidationError(err.Error()))
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Output Validation
```go
// Response validation middleware
func ValidateResponse(schema string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            rw := newResponseWriter(w)
            
            next.ServeHTTP(rw, r)
            
            if err := validator.ValidateJSON(schema, rw.body.Bytes()); err != nil {
                // Log validation error but send original response
                log.Printf("Response validation failed: %v", err)
            }
        })
    }
}
```

### Process Configuration Validation
```go
func (p *Process) ValidateConfig(config map[string]interface{}) error {
    schema, err := p.GetConfigSchema()
    if err != nil {
        return fmt.Errorf("getting config schema: %w", err)
    }
    
    if err := validator.Validate(schema, config); err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }
    
    return nil
}
```

## Common Schemas

### Flow Configuration Schema
```json
{
    "type": "object",
    "required": ["id", "nodes", "edges"],
    "properties": {
        "id": {
            "type": "string",
            "pattern": "^[a-zA-Z0-9-_]+$"
        },
        "nodes": {
            "type": "object",
            "additionalProperties": {
                "$ref": "#/definitions/NodeConfig"
            }
        },
        "edges": {
            "type": "array",
            "items": {
                "$ref": "#/definitions/EdgeConfig"
            }
        }
    }
}
```

### Node Configuration Schema
```json
{
    "type": "object",
    "required": ["type", "config"],
    "properties": {
        "type": {
            "type": "string"
        },
        "config": {
            "type": "object"
        }
    }
}
```

## Best Practices

### Input Validation
- Validate all external input
- Use strict schema validation
- Sanitize string inputs
- Check numerical bounds
- Validate array lengths
- Verify required fields

### Type Safety
- Use Go's type system
- Implement custom unmarshaling
- Handle type conversions explicitly
- Validate complex types
- Check enum values

### Error Handling
- Return clear error messages
- Include validation context
- Log validation failures
- Handle partial validation
- Support custom errors

### Performance
- Cache compiled schemas
- Use efficient validators
- Validate at edges only
- Handle large inputs
- Consider validation cost

## Testing

### Schema Testing
```go
func TestValidation(t *testing.T) {
    testCases := []struct{
        name    string
        schema  string
        input   interface{}
        wantErr bool
    }{
        {
            name: "valid input",
            schema: flowConfigSchema,
            input: map[string]interface{}{
                "id": "test-flow",
                "nodes": map[string]interface{}{},
                "edges": []interface{}{},
            },
            wantErr: false,
        },
        // Add more test cases...
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := validator.Validate(tc.schema, tc.input)
            if (err != nil) != tc.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
            }
        })
    }
}
```

### Integration Testing
```go
func TestValidationMiddleware(t *testing.T) {
    handler := ValidateRequest(testSchema)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    
    srv := httptest.NewServer(handler)
    defer srv.Close()
    
    // Test valid request
    resp, err := http.Post(srv.URL, "application/json", strings.NewReader(`{"valid":"data"}`))
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // Test invalid request
    resp, err = http.Post(srv.URL, "application/json", strings.NewReader(`{"invalid":[]}`))
    require.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
```

