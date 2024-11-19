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
        }
    },
    "edges": []
}
```

**Responses:**
- 201 Created: Flow created successfully
  ```json
  {
      "data": {
          "id": "example-flow",
          "config": {
              "id": "example-flow",
              "nodes": {
                  "reader": {
                      "type": "FileReader",
                      "config": {
                          "filename": "input.txt"
                      }
                  }
              },
              "edges": []
          },
          "state": "created",
          "started_at": null,
          "error": ""
      }
  }
  ```
- 400 Bad Request: Invalid configuration
  ```json
  {
      "error": {
          "message": "invalid process type: InvalidType"
      }
  }
  ```
- 409 Conflict: Flow already exists
  ```json
  {
      "error": {
          "message": "flow example-flow already exists"
      }
  }
  ```

### Get Flow
```http
GET /api/v1/flows/{id}
```

**Responses:**
- 200 OK: Flow found
  ```json
  {
      "data": {
          "id": "example-flow",
          "config": {
              // Flow configuration
          },
          "state": "running",
          "started_at": "2024-01-01T12:00:00Z",
          "error": ""
      }
  }
  ```
- 404 Not Found: Flow does not exist
  ```json
  {
      "error": {
          "message": "flow example-flow not found"
      }
  }
  ```

### List Flows
```http
GET /api/v1/flows
```

**Response:**
```json
{
    "data": [
        {
            "id": "flow-1",
            "config": {
                // Flow configuration
            },
            "state": "running",
            "started_at": "2024-01-01T12:00:00Z",
            "error": ""
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
  ```json
  {
      "data": {
          "id": "example-flow",
          "state": "starting",
          "started_at": "2024-01-01T12:00:00Z"
      }
  }
  ```
- 404 Not Found: Flow not found
- 409 Conflict: Flow already running

### Stop Flow
```http
POST /api/v1/flows/{id}/stop
```

**Responses:**
- 200 OK: Flow stopped successfully
  ```json
  {
      "data": {
          "id": "example-flow",
          "state": "stopping"
      }
  }
  ```
- 404 Not Found: Flow not found
- 409 Conflict: Flow not running

### Delete Flow
```http
DELETE /api/v1/flows/{id}
```

**Responses:**
- 204 No Content: Flow deleted successfully
- 404 Not Found: Flow not found
- 409 Conflict: Cannot delete running flow

### Flow States
The following states are supported:
- `created`: Initial state after flow creation
- `starting`: Flow is in the process of starting
- `running`: Flow is actively running
- `stopping`: Flow is in the process of stopping
- `stopped`: Flow has been stopped
- `error`: Flow encountered an error

## Error Responses
All error responses follow the format:
```json
{
    "error": {
        "message": "Human readable error message"
    }
}
```

Common HTTP status codes:
- 400 Bad Request: Invalid request or configuration
- 404 Not Found: Resource not found
- 409 Conflict: Resource state conflict
- 500 Internal Server Error: Server error

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