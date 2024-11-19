# Documentation Server

This document details how documentation and visualizations are served through the Flow Server.

## Core Components

### Documentation Server
```go
type DocsServer struct {
    router     *chi.Router
    docsPath   string
    mermaidGen *MermaidGenerator
}

func NewDocsServer(config DocsConfig) *DocsServer {
    return &DocsServer{
        router:     chi.NewRouter(),
        docsPath:   config.DocsPath,
        mermaidGen: NewMermaidGenerator(),
    }
}
```

### Documentation Routes
```go
func (s *DocsServer) setupRoutes() {
    // Static documentation
    s.router.Get("/docs/*", http.StripPrefix("/docs/", 
        http.FileServer(http.Dir(s.docsPath))))
    
    // Mermaid diagrams
    s.router.Get("/diagrams/network/{id}", s.handleNetworkDiagram)
    s.router.Get("/diagrams/network/{id}/live", s.handleLiveDiagram)
    
    // API documentation
    s.router.Get("/api-docs", s.handleSwaggerUI)
}
```

## Features

### Live Documentation
- Markdown rendering
- Code syntax highlighting
- Navigation sidebar
- Search functionality

### Interactive Diagrams
- Network visualization
- Live updates
- Status indicators
- Zoom and pan

## Integration

### Flow Server Integration
```go
type FlowServer struct {
    // ... other fields ...
    docsServer *DocsServer
}

func (s *FlowServer) setupDocs() {
    s.router.Mount("/docs", s.docsServer.router)
}
```

### Network Visualization
```go
func (s *DocsServer) handleNetworkDiagram(w http.ResponseWriter, r *http.Request) {
    networkID := chi.URLParam(r, "id")
    
    diagram, err := s.mermaidGen.GenerateFlowDiagram(networkID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    s.renderDiagram(w, diagram)
}
```

## Usage

### Starting the Server
```go
server := NewFlowServer(FlowServerConfig{
    Port: 8080,
    DocsConfig: DocsConfig{
        DocsPath: "./docs",
        EnableLive: true,
    },
})

if err := server.Run(); err != nil {
    log.Fatal(err)
}
```

### Accessing Documentation
```bash
# View documentation
open http://localhost:8080/docs

# View network diagram
open http://localhost:8080/diagrams/network/flow1

# View API documentation
open http://localhost:8080/api-docs
``` 