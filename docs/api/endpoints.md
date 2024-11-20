# API Documentation

## Core Endpoints

### Documentation Endpoints

#### Static Documentation
```http
GET /docs/*
```
Serves static documentation files from configured docs directory
- Markdown files served with HTML wrapper
- API documentation
- README.md as home page

#### API Documentation
```http
GET /api/*
```
Serves API documentation from docs/api directory
- Swagger/OpenAPI specification
- API schemas
- API documentation

#### Home Page
```http
GET /
```
Serves README.md as the home page

### Flow Management

#### Create Flow
```http
POST /flows
```
Request:
```json
{
    "id": "flow-id",
    "config": {
        "nodes": {
            "node-id": {
                "type": "ProcessType",
                "config": {}
            }
        }
    }
}
```

#### Get Flow
```http
GET /flows/{id}
```
Response:
```json
{
    "id": "flow-id",
    "state": "created|running|stopped",
    "config": {}
}
```

#### Start Flow
```http
POST /flows/{id}/start
```

#### Stop Flow
```http
POST /flows/{id}/stop
```

## Response Formats

### Success Response
```json
{
    "data": {
        // Response data
    }
}
```

### Error Response
```json
{
    "error": {
        "message": "Error description"
    }
}
```
