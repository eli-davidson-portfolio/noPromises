# Server Configuration

This document details the configuration options for the noPromises server.

## Core Configuration

### Server Config
```go
type Config struct {
    Port     int    // HTTP server port
    DocsPath string // Path to documentation files
}
```

### Documentation Config
```go
type DocsConfig struct {
    DocsPath    string // Documentation root directory
    EnableLive  bool   // Enable live updates
    SwaggerPath string // Path to swagger.json
}
```

## Configuration Validation

### Path Validation
```go
func validateConfig(config Config) error {
    // Verify required files exist
    requiredFiles := []string{
        "README.md",
        "api/swagger.json",
    }
    
    for _, file := range requiredFiles {
        path := filepath.Join(config.DocsPath, file)
        if _, err := os.Stat(path); os.IsNotExist(err) {
            return fmt.Errorf("required file missing: %s", path)
        }
    }
    
    return nil
}
```

### Permission Checks
```go
func checkPermissions(path string) error {
    // Check read permissions
    if _, err := os.ReadFile(filepath.Join(path, "README.md")); err != nil {
        return fmt.Errorf("insufficient permissions: %v", err)
    }
    
    return nil
}
```

## Usage Examples

### Basic Configuration
```go
config := Config{
    Port:     8080,
    DocsPath: "./docs",
}

server, err := NewServer(config)
if err != nil {
    log.Fatal(err)
}
```

### Custom Configuration
```go
config := Config{
    Port:     3000,
    DocsPath: "/var/www/docs",
}

docsConfig := DocsConfig{
    DocsPath:    config.DocsPath,
    EnableLive:  true,
    SwaggerPath: "/api/swagger.json",
}

server := NewServer(config, WithDocs(docsConfig))
```

## Directory Structure

### Required Structure
```
docs/
├── api/              # API documentation
│   └── swagger.json  # OpenAPI specification
├── architecture/     # Architecture documentation
├── guides/          # User guides
└── README.md        # Documentation home
```

### File Requirements
1. `README.md` - Main documentation entry point
2. `api/swagger.json` - OpenAPI specification
3. Directory structure for organization
4. Proper file permissions

## Best Practices

### Path Configuration
- Use absolute paths in production
- Validate paths before server start
- Set appropriate permissions
- Include all required sections

### Port Configuration
- Use standard ports (8080, 3000) for development
- Configure firewall rules appropriately
- Handle port conflicts gracefully
- Log port binding errors

### Documentation Setup
- Organize files logically
- Maintain consistent structure
- Include required files
- Set correct permissions

### Security Considerations
- Validate all paths
- Check file permissions
- Prevent directory traversal
- Secure sensitive docs

## Error Handling

### Configuration Errors
```go
type ConfigError struct {
    Field   string
    Message string
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("configuration error: %s - %s", e.Field, e.Message)
}
```

### Common Error Cases
- Missing required files
- Invalid permissions
- Port already in use
- Invalid paths

## Monitoring

### Configuration Checks
- Log configuration loading
- Monitor file access
- Track permission issues
- Report path problems

### Health Checks
- Verify file access
- Check port availability
- Monitor permissions
- Track configuration changes
