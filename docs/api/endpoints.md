# API Endpoints

## Authentication

### Generate Token
```http
POST /token
Content-Type: application/json

{
    "username": "string",
    "password": "string"
}
```

Response:
```json
{
    "token": "jwt.token.here"
}
```

### Protected Endpoints
All endpoints except `/health` and `/token` require authentication via Bearer token:

```http
GET /protected-endpoint
Authorization: Bearer <token>
```

## Health Check

```http
GET /health
```

Returns 501 Not Implemented if no router is configured.

## Error Responses

### Unauthorized
```json
{
    "error": "missing authorization token"
}
```

### Rate Limited
```json
{
    "error": "rate limit exceeded"
}
```

## Testing

The API includes test endpoints that verify:
- Basic API functionality
- Authentication flow
- Rate limiting
- Concurrent operations
