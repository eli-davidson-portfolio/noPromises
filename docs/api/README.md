# noPromises API Documentation

Welcome to the noPromises API documentation. This documentation provides comprehensive information about the API endpoints, schemas, and usage.

## Contents

- [API Endpoints](endpoints.md) - Detailed documentation of all available API endpoints
- [JSON Schemas](schemas.md) - JSON schemas for request and response payloads
- [OpenAPI Specification](swagger.json) - Complete OpenAPI/Swagger specification

## Quick Start

### Core Endpoints

1. **Flow Management**
   - Create Flow: `POST /flows`
   - Get Flow: `GET /flows/{id}`
   - Start Flow: `POST /flows/{id}/start`
   - Stop Flow: `POST /flows/{id}/stop`

2. **Documentation**
   - Static Docs: `GET /docs/*`
   - API Docs: `GET /api/*`
   - Home Page: `GET /`

3. **Diagrams**
   - Network Diagram: `GET /diagrams/network/{id}`
   - Live Updates: `GET /diagrams/network/{id}/live` (WebSocket)

### Response Format

Success Response:
```json
{
    "data": {
        // Response payload
    }
}
```

Error Response:
```json
{
    "error": {
        "message": "Error description"
    }
}
```

## Interactive Documentation

For interactive API documentation, visit:
- Swagger UI: `/api-docs`
- OpenAPI Specification: `/api/swagger.json`

## Further Reading

- For detailed endpoint documentation, see [endpoints.md](endpoints.md)
- For JSON schema definitions, see [schemas.md](schemas.md)
- For complete API specification, see [swagger.json](swagger.json)
