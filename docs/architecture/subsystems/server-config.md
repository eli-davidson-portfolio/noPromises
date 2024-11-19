# Server Configuration

This document details the configuration options for the noPromises server.

## Core Configuration

```go
// Config holds server configuration
type Config struct {
    Port     int    // HTTP server port
    DocsPath string // Path to documentation files
}
```

## Usage

### Command Line Flags
```go
func main() {
    port := flag.Int("port", 8080, "Server port")
    docsPath := flag.String("docs", "./docs", "Path to documentation files")
    flag.Parse()
    
    srv, err := server.NewServer(server.Config{
        Port:     *port,
        DocsPath: *docsPath,
    })
}
```

## Documentation Configuration

The documentation server requires:
1. Valid path to documentation files
2. Markdown files in appropriate directories
3. Proper file permissions for reading

## Best Practices

### Documentation Path
- Use absolute paths in production
- Validate path exists before starting
- Set appropriate permissions
- Include all required documentation sections

### Port Configuration
- Use standard ports (8080, 3000) for development
- Configure firewall rules appropriately
- Handle port conflicts gracefully
- Log port binding errors
