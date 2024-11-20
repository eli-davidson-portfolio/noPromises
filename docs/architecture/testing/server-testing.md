# Server Testing Architecture

## Test Setup

### Test Server Setup
```go
func setupTestServer(t *testing.T) *server.Server {
    // Create temporary test directories
    tmpDir := t.TempDir()
    docsDir := filepath.Join(tmpDir, "docs")
    webDir := filepath.Join(tmpDir, "web", "templates")
    
    // Create required directories
    require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "api"), 0755))
    require.NoError(t, os.MkdirAll(webDir, 0755))

    // Create required files
    require.NoError(t, os.WriteFile(
        filepath.Join(docsDir, "README.md"),
        []byte("# Test Documentation"),
        0644,
    ))
    require.NoError(t, os.WriteFile(
        filepath.Join(docsDir, "api", "swagger.json"),
        []byte("{}"),
        0644,
    ))

    // Create server with test configuration
    srv, err := server.NewServer(server.Config{
        Port:     0, // Use random port
        DocsPath: docsDir,
    })
    require.NoError(t, err)
    require.NotNil(t, srv)

    return srv
}
```

## Mock Implementations

### Mock Process
```go
type mockProcess struct {
    id     string
    status string
}

func (m *mockProcess) Start(_ context.Context) error {
    m.status = "running"
    return nil
}

func (m *mockProcess) Stop(_ context.Context) error {
    m.status = "stopped"
    return nil
}

func (m *mockProcess) ID() string {
    return m.id
}
```

### Mock Flow Manager
```go
type mockFlowManager struct {
    flows []ManagedFlow
}

func (m *mockFlowManager) List() []ManagedFlow {
    return m.flows
}

func (m *mockFlowManager) Get(id string) (*ManagedFlow, bool) {
    for _, flow := range m.flows {
        if flow.ID == id {
            return &flow, true
        }
    }
    return nil, false
}
```

## Test Categories

### Server Tests
- Server initialization
- Configuration validation
- Route setup
- Server lifecycle (start/stop)

### Web Interface Tests
- Home page rendering
- Flow list display
- Documentation serving
- API documentation access

### Flow Management Tests
- Flow creation
- Flow retrieval
- Flow state management
- Flow lifecycle

### Process Tests
- Process type registration
- Process creation
- Process lifecycle
- Error handling

## Integration Tests

### Web Integration
```go
func TestWebIntegration(t *testing.T) {
    srv := setupTestServer(t)

    t.Run("serve home page", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/", nil)
        w := httptest.NewRecorder()
        srv.ServeHTTP(w, req)
        assert.Equal(t, http.StatusOK, w.Code)
        assert.Contains(t, w.Body.String(), "Test Documentation")
    })

    t.Run("serve docs", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/docs/README.md", nil)
        w := httptest.NewRecorder()
        srv.ServeHTTP(w, req)
        assert.Equal(t, http.StatusOK, w.Code)
    })
}
```

## Best Practices

### Test Setup
- Use t.TempDir() for temporary files
- Clean up resources after tests
- Use random ports for servers
- Isolate test environments

### Mock Usage
- Keep mocks simple
- Mock at interface boundaries
- Verify mock interactions
- Document mock behavior

### Test Organization
- Group related tests
- Use table-driven tests
- Test error cases
- Test concurrent operations

### Test Coverage
- Core functionality
- Error handling
- Edge cases
- Resource cleanup