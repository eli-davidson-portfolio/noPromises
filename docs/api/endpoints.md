# API Documentation

This document details the noPromises API endpoints.

## Documentation Endpoints

### Static Documentation
```http
GET /docs/*
```
Serves Markdown documentation files with HTML wrapper
- Automatically converts Markdown to styled HTML
- Includes syntax highlighting for code blocks
- Responsive design with mobile support
- Navigation sidebar

### API Documentation
```http
GET /api-docs
```
Serves Swagger UI interface
- Interactive API documentation
- Try-it-now functionality
- Schema visualization
- Request/response examples

### OpenAPI Specification
```http
GET /api/swagger.json
```
Serves OpenAPI/Swagger specification
- Complete API schema
- Endpoint definitions
- Data models
- Authentication requirements

## Network Visualization

### Get Network Diagram
```http
GET /diagrams/network/{id}

Response:
{
    "diagram": "graph LR\n    reader[FileReader]:::running\n    writer[FileWriter]:::waiting\n    reader -->|data| writer"
}
```
Generates Mermaid diagram for network visualization
- Node status indication
- Connection visualization
- Port labeling
- State-based styling

### Live Network Updates
```http
GET /diagrams/network/{id}/live
```
WebSocket endpoint for real-time updates
- Live state changes
- Connection status
- Error indication
- Performance metrics

## Response Formats

### Success Response
```json
{
    "data": {
        // Response data specific to endpoint
    }
}
```

### Error Response
```json
{
    "error": {
        "message": "Error description",
        "code": "ERROR_CODE",
        "details": {}
    }
}
```

## Common Status Codes
- 200: Successful request
- 201: Resource created
- 400: Bad request
- 404: Resource not found
- 500: Server error
- 101: Switching protocols (WebSocket)

## Content Types
- `application/json`: API responses
- `text/html`: Documentation pages
- `text/markdown`: Raw documentation
- `application/json`: Diagram data

## Authentication
Currently using basic authentication for all endpoints
- Include credentials in request header
- Token-based auth coming soon
- Rate limiting applied to all endpoints
