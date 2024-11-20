# Server Testing Guide

This document outlines testing patterns for the Flow Server subsystem.

## Documentation Server Tests

### Server Creation
```go
func TestDocsServer(t *testing.T) {
    // Create test docs directory
    tmpDir := t.TempDir()
    require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "api"), 0755))
    
    // Create test files
    require.NoError(t, os.WriteFile(
        filepath.Join(tmpDir, "test.md"),
        []byte("# Test Documentation"),
        0644,
    ))
    
    // Create server
    srv := NewServer(Config{
        DocsPath: tmpDir,
    })
    require.NotNil(t, srv)
}
```

### Route Testing
```go
func TestRoutes(t *testing.T) {
    srv := setupTestServer(t)
    
    tests := []struct {
        name           string
        path           string
        expectedStatus int
        expectedType   string
        checkResponse  func(*testing.T, *httptest.ResponseRecorder)
    }{
        {
            name: "serve markdown",
            path: "/test.md",
            expectedStatus: http.StatusOK,
            expectedType: "text/html; charset=utf-8",
            checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
                assert.Contains(t, w.Body.String(), "<html>")
                assert.Contains(t, w.Body.String(), "Test Documentation")
            },
        },
        {
            name: "serve swagger",
            path: "/api/swagger.json",
            expectedStatus: http.StatusOK,
            expectedType: "application/json",
            checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
                assert.Contains(t, w.Body.String(), `"openapi":"3.0.0"`)
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", tt.path, nil)
            w := httptest.NewRecorder()
            
            srv.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
            assert.Contains(t, w.Header().Get("Content-Type"), tt.expectedType)
            
            if tt.checkResponse != nil {
                tt.checkResponse(t, w)
            }
        })
    }
}
```

### Network Visualization Tests
```go
func TestNetworkDiagram(t *testing.T) {
    srv := setupTestServer(t)
    
    // Set test network
    srv.mermaidGen.SetNetwork("test-flow", map[string]interface{}{
        "nodes": map[string]interface{}{
            "reader": map[string]interface{}{
                "type": "FileReader",
                "status": "running",
            },
        },
    })
    
    // Test diagram generation
    req := httptest.NewRequest("GET", "/diagrams/network/test-flow", nil)
    w := httptest.NewRecorder()
    
    srv.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "reader[FileReader]:::running")
}
```

### WebSocket Tests
```go
func TestLiveUpdates(t *testing.T) {
    srv := setupTestServer(t)
    
    req := httptest.NewRequest("GET", "/diagrams/network/test-flow/live", nil)
    w := httptest.NewRecorder()
    
    srv.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusSwitchingProtocols, w.Code)
}
```

## Test Utilities

### Server Setup
```go
func setupTestServer(t *testing.T) *Server {
    tmpDir := t.TempDir()
    
    // Create required files
    setupTestFiles(t, tmpDir)
    
    srv := NewServer(Config{
        DocsPath: tmpDir,
    })
    srv.SetupRoutes()
    
    return srv
}
```

### Test File Creation
```go
func setupTestFiles(t *testing.T, dir string) {
    files := map[string]string{
        "README.md": "# Test Documentation",
        "api/swagger.json": `{"openapi":"3.0.0"}`,
        "test.md": "# Test Content",
    }
    
    for path, content := range files {
        fullPath := filepath.Join(dir, path)
        require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
        require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
    }
}
```

## Best Practices

### Test Organization
- Group related tests
- Use table-driven tests
- Isolate test dependencies
- Clean up test resources

### Error Testing
- Test missing files
- Test invalid paths
- Test malformed content
- Test error responses

### Content Testing
- Verify HTML wrapping
- Check content types
- Validate JSON responses
- Test diagram generation

### Integration Testing
- Test full request flow
- Verify content serving
- Check live updates
- Test error handling

## Database Testing

### Test Setup
```go
func setupTestDB(t *testing.T) *db.SQLiteDB {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    db, err := db.NewSQLiteDB(dbPath)
    require.NoError(t, err)
    require.NotNil(t, db)

    t.Cleanup(func() {
        db.Close()
        os.Remove(dbPath)
        os.Remove(dbPath + "-shm") // Remove WAL files
        os.Remove(dbPath + "-wal")
    })

    return db
}
```

### Database Tests
```go
func TestDatabaseOperations(t *testing.T) {
    db := setupTestDB(t)

    // Test WAL mode
    var journalMode string
    err := db.DB().QueryRow("PRAGMA journal_mode").Scan(&journalMode)
    require.NoError(t, err)
    assert.Equal(t, "wal", journalMode)

    // Test foreign keys
    var foreignKeys int
    err = db.DB().QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
    require.NoError(t, err)
    assert.Equal(t, 1, foreignKeys)
}
```