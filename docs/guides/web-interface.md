# Web Interface Documentation

## HTML Templates

### Documentation Template
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>noPromises Documentation</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/github-markdown-css@5/github-markdown.min.css">
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

## Flow Management UI

### Network Visualization
- Real-time network state display
- Interactive node inspection
- Connection status visualization
- Performance metrics display

### Diagram Features
```javascript
// Mermaid diagram configuration
const config = {
    theme: 'default',
    flowchart: {
        curve: 'basis',
        nodeSpacing: 50,
        rankSpacing: 50
    }
}
```

### State Indicators
- Running: Green background
- Waiting: Yellow background
- Error: Red background
- Inactive: Gray background

## Live Updates

### WebSocket Integration
```javascript
const ws = new WebSocket(`ws://${window.location.host}/diagrams/network/${flowId}/live`)
ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    updateDiagram(data.diagram)
}
```

### Update Handling
```javascript
function updateDiagram(diagram) {
    // Clear existing diagram
    document.getElementById('diagram').innerHTML = ''
    
    // Render new diagram
    mermaid.render('diagram-svg', diagram, (svg) => {
        document.getElementById('diagram').innerHTML = svg
    })
}
```

## Responsive Design

### Mobile Support
- Adaptive layout
- Touch-friendly controls
- Responsive diagrams
- Optimized performance

### Desktop Features
- Extended controls
- Detailed metrics
- Multi-pane view
- Advanced filtering

## Integration Points

### API Documentation
- Swagger UI integration
- Interactive API testing
- Schema visualization
- Request/response examples

### Network Management
- Flow creation interface
- Process configuration
- Connection management
- State control

## Styling Guidelines

### Theme Configuration
```css
:root {
    --primary-color: #0366d6;
    --error-color: #dc3545;
    --success-color: #28a745;
    --warning-color: #ffc107;
}
```

### Component Styles
- Consistent padding/margins
- Clear visual hierarchy
- State-based colors
- Responsive breakpoints