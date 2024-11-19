# Documentation Server

This document details how documentation and visualizations are served through the Flow Server.

## Core Components

### Documentation Server
```go
type Server struct {
    router     *mux.Router
    docsPath   string
    mermaidGen *MermaidGenerator
}

func NewServer(config Config) *Server {
    return &Server{
        router:     mux.NewRouter(),
        docsPath:   config.DocsPath,
        mermaidGen: NewMermaidGenerator(),
    }
}
```

### Route Setup
```go
func (s *Server) SetupRoutes() {
    // Documentation routes
    s.router.PathPrefix("/docs/").Handler(s.wrapMarkdown(http.FileServer(http.Dir(s.docsPath))))
    
    // API documentation
    s.router.HandleFunc("/api-docs", s.HandleSwaggerUI)
    s.router.HandleFunc("/api/swagger.json", s.serveSwaggerJSON)
    
    // Network visualization
    s.router.HandleFunc("/diagrams/network/{id}", s.handleNetworkDiagram)
    s.router.HandleFunc("/diagrams/network/{id}/live", s.handleLiveDiagram)
}
```

## Features

### Markdown Processing
```go
func (s *Server) wrapMarkdown(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasSuffix(r.URL.Path, ".md") {
            content, err := os.ReadFile(filepath.Join(s.docsPath, r.URL.Path))
            if err != nil {
                http.Error(w, "Documentation not found", http.StatusNotFound)
                return
            }
            s.renderDocPage(w, string(content))
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### HTML Wrapper
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>noPromises Documentation</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/github-markdown-css@5/github-markdown.min.css">
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
</head>
<body>
    <nav>
        <a href="/docs">Home</a>
        <a href="/docs/guides">Guides</a>
        <a href="/api-docs">API</a>
    </nav>
    <div class="markdown-body">
        <!-- Content inserted here -->
    </div>
</body>
</html>
```

### Network Visualization
```go
func (s *Server) handleNetworkDiagram(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    networkID := vars["id"]

    diagram, err := s.mermaidGen.GenerateFlowDiagram(networkID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "diagram": diagram,
    })
}
```

## Integration

### Server Integration
```go
func (s *Server) Router() *mux.Router {
    return s.router
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.router.ServeHTTP(w, r)
}
```

### Debug Logging
```go
func (s *Server) logDebug(format string, args ...interface{}) {
    log.Printf("[DEBUG] "+format, args...)
}
```

## Best Practices

### Documentation Organization
- Keep documentation close to code
- Use consistent file structure
- Follow Markdown conventions
- Include code examples

### Error Handling
- Provide clear error messages
- Log debugging information
- Return appropriate status codes
- Handle missing files gracefully

### Security
- Validate file paths
- Sanitize user input
- Control access to sensitive docs
- Log access attempts

### Performance
- Cache rendered content
- Optimize large files
- Handle concurrent requests
- Monitor resource usage
``` 
</rewritten_file>