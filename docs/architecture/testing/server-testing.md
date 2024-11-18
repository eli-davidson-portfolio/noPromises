# Server Testing Guide

This document outlines testing patterns for the Flow Server subsystem.

## Unit Tests

### Server Creation
```go
func TestFlowServer(t *testing.T) {
    t.Run("creation", func(t *testing.T) {
        config := ServerConfig{
            Port: 8080,
        }
        
        server := NewFlowServer(config)
        assert.NotNil(t, server)
        assert.Equal(t, 8080, server.port)
    })
}
```

### Process Registry
```go
func TestProcessRegistry(t *testing.T) {
    t.Run("registration", func(t *testing.T) {
        registry := NewProcessRegistry()
        
        err := registry.Register("test", func(config ProcessConfig) (Process, error) {
            return NewTestProcess(), nil
        })
        require.NoError(t, err)
        
        factory, exists := registry.Get("test")
        assert.True(t, exists)
        assert.NotNil(t, factory)
    })
}
```

### Flow Management
```go
func TestFlowManagement(t *testing.T) {
    t.Run("create flow", func(t *testing.T) {
        server := NewFlowServer(ServerConfig{})
        
        config := FlowConfig{
            ID: "test-flow",
            Nodes: map[string]NodeConfig{
                "proc1": {
                    Type: "test",
                    Config: map[string]interface{}{},
                },
            },
        }
        
        err := server.CreateFlow(config)
        require.NoError(t, err)
        
        assert.Equal(t, 1, server.ProcessCount())
    })
}
```

## Integration Tests

### HTTP Endpoints
```go
func TestHTTPEndpoints(t *testing.T) {
    t.Run("create flow", func(t *testing.T) {
        server := NewTestServer()
        
        resp := httptest.NewRecorder()
        req := httptest.NewRequest("POST", "/api/flows", strings.NewReader(`{
            "id": "test-flow",
            "nodes": {
                "proc1": {
                    "type": "test",
                    "config": {}
                }
            }
        }`))
        
        server.ServeHTTP(resp, req)
        assert.Equal(t, http.StatusCreated, resp.Code)
    })
}
```

### Flow Execution
```go
func TestFlowExecution(t *testing.T) {
    t.Run("start and stop", func(t *testing.T) {
        server := NewTestServer()
        
        // Create flow
        config := FlowConfig{
            ID: "test-flow",
            Nodes: map[string]NodeConfig{
                "proc1": {Type: "test"},
            },
        }
        require.NoError(t, server.CreateFlow(config))
        
        // Start flow
        err := server.StartFlow("test-flow")
        require.NoError(t, err)
        
        // Verify status
        status, err := server.GetFlowStatus("test-flow")
        require.NoError(t, err)
        assert.Equal(t, NetworkStatusRunning, status)
        
        // Stop flow
        err = server.StopFlow("test-flow")
        require.NoError(t, err)
        
        // Verify stopped
        status, err = server.GetFlowStatus("test-flow")
        require.NoError(t, err)
        assert.Equal(t, NetworkStatusStopped, status)
    })
}
```

## Concurrency Tests

### Parallel Operations
```go
func TestConcurrentOperations(t *testing.T) {
    t.Run("parallel flow creation", func(t *testing.T) {
        server := NewTestServer()
        
        var wg sync.WaitGroup
        for i := 0; i < 10; i++ {
            wg.Add(1)
            go func(id int) {
                defer wg.Done()
                config := FlowConfig{
                    ID: fmt.Sprintf("flow-%d", id),
                    Nodes: map[string]NodeConfig{
                        "proc1": {Type: "test"},
                    },
                }
                err := server.CreateFlow(config)
                assert.NoError(t, err)
            }(i)
        }
        wg.Wait()
        
        assert.Equal(t, 10, server.ProcessCount())
    })
}
```

### Race Condition Tests
```go
func TestRaceConditions(t *testing.T) {
    t.Run("concurrent state access", func(t *testing.T) {
        server := NewTestServer()
        
        // Create test flow
        config := FlowConfig{ID: "test-flow"}
        require.NoError(t, server.CreateFlow(config))
        
        // Concurrent operations
        var wg sync.WaitGroup
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                _, _ = server.GetFlowStatus("test-flow")
            }()
        }
        wg.Wait()
    })
}
```

## Best Practices

### Test Setup
- Use test helpers
- Clean up resources
- Isolate tests
- Handle timeouts

### Error Testing
- Test error cases
- Verify error messages
- Check status codes
- Test cleanup

### Concurrency Testing
- Use race detector
- Test parallel operations
- Check state consistency
- Verify thread safety

### Integration Testing
- Test real HTTP
- Verify responses
- Check headers
- Test timeouts 