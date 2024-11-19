# API Endpoints

## Overview

The noPromises server exposes a RESTful HTTP API for managing Flow-Based Programming (FBP) networks. All endpoints are prefixed with `/api/v1`.

## Flow Management

### Create Flow
```http
POST /api/v1/flows
Content-Type: application/json

{
    "id": "example-flow",
    "nodes": {
        "reader": {
            "type": "FileReader",
            "config": {
                "filename": "input.txt"
            }
        },
        "writer": {
            "type": "FileWriter",
            "config": {
                "filename": "output.txt"
            }
        }
    },
    "edges": [
        {
            "from": {"node": "reader", "port": "out"},
            "to": {"node": "writer", "port": "in"}
        }
    ]
}
```

**Responses:**
- 201 Created: Flow created successfully
  ```json
  {
      "id": "example-flow",
      "status": "created",
      "created_at": "2024-11-19T10:00:00Z"
  }
  ```
- 400 Bad Request: Invalid configuration
- 409 Conflict: Flow ID already exists

### List Flows
```http
GET /api/v1/flows
```

**Response:**
```json
{
    "flows": [
        {
            "id": "example-flow",
            "state": "running",
            "node_count": 2,
            "created_at": "2024-11-19T10:00:00Z"
        }
    ]
}
```

### Get Flow Details
```http
GET /api/v1/flows/{id}
```

**Response:**
```json
{
    "id": "example-flow",
    "state": "running",
    "started_at": "2024-11-19T10:00:00Z",
    "nodes": {
        "reader": {
            "type": "FileReader",
            "config": {
                "filename": "input.txt"
            },
            "status": {
                "state": "running",
                "messages_processed": 150,
                "last_active": "2024-11-19T10:05:00Z"
            }
        }
    },
    "edges": [
        {
            "from": {"node": "reader", "port": "out"},
            "to": {"node": "writer", "port": "in"}
        }
    ]
}
```

### Start Flow
```http
POST /api/v1/flows/{id}/start
```

**Responses:**
- 200 OK: Flow started successfully
- 404 Not Found: Flow not found
- 409 Conflict: Flow already running

### Stop Flow
```http
POST /api/v1/flows/{id}/stop
```

**Responses:**
- 200 OK: Flow stopped successfully
- 404 Not Found: Flow not found
- 409 Conflict: Flow not running

### Delete Flow
```http
DELETE /api/v1/flows/{id}
```

**Responses:**
- 204 No Content: Flow deleted successfully
- 404 Not Found: Flow not found
- 409 Conflict: Flow currently running

### Get Flow Status
```http
GET /api/v1/flows/{id}/status
```

**Response:**
```json
{
    "id": "example-flow",
    "state": "running",
    "started_at": "2024-11-19T10:00:00Z",
    "nodes": {
        "reader": {
            "state": "running",
            "messages_processed": 150,
            "last_active": "2024-11-19T10:05:00Z"
        }
    }
}
```

## Process Discovery

### List Available Processes
```http
GET /api/v1/processes
```

**Response:**
```json
{
    "processes": [
        {
            "name": "FileReader",
            "description": "Reads data from a file",
            "input_ports": [],
            "output_ports": [
                {
                    "name": "out",
                    "description": "Data read from file",
                    "required": true
                }
            ],
            "config": {
                "properties": {
                    "filename": {
                        "type": "string",
                        "description": "Path to input file",
                        "required": true
                    }
                }
            }
        }
    ]
}
```

### Get Process Details
```http
GET /api/v1/processes/{name}
```

**Response:**
```json
{
    "name": "FileReader",
    "description": "Reads data from a file",
    "input_ports": [],
    "output_ports": [
        {
            "name": "out",
            "description": "Data read from file",
            "required": true
        }
    ],
    "config": {
        "properties": {
            "filename": {
                "type": "string",
                "description": "Path to input file",
                "required": true
            }
        }
    }
}
```

## Error Responses

All error responses follow the format:

```json
{
    "error": {
        "code": "FLOW_NOT_FOUND",
        "message": "Flow 'example-flow' not found",
        "details": {
            "flow_id": "example-flow"
        }
    }
}
```

Common error codes:
- `INVALID_REQUEST`: Malformed request or invalid parameters
- `FLOW_NOT_FOUND`: Requested flow does not exist
- `FLOW_EXISTS`: Flow ID already in use
- `INVALID_STATE`: Invalid flow state for operation
- `PROCESS_NOT_FOUND`: Process type does not exist
- `VALIDATION_ERROR`: Configuration validation failed
- `INTERNAL_ERROR`: Unexpected server error

## Request Headers

Required headers:
- `Content-Type: application/json` for all POST/PUT requests

Optional headers:
- `Accept: application/json` (default if not specified)
- `X-Request-ID`: Client-provided request identifier

## Rate Limiting

The API implements rate limiting per client IP address:

- Rate limit headers included in all responses:
  ```
  X-RateLimit-Limit: 100
  X-RateLimit-Remaining: 99
  X-RateLimit-Reset: 1573664000
  ```

- When limit exceeded, returns 429 Too Many Requests:
  ```json
  {
      "error": {
          "code": "RATE_LIMIT_EXCEEDED",
          "message": "Rate limit exceeded. Try again in 60 seconds",
          "details": {
              "retry_after": 60
          }
      }
  }
  ```

## Versioning

All endpoints are versioned:
- Current version: v1
- Version included in URL path: `/api/v1/`
- Version header included in all responses: `X-API-Version: 1.0`