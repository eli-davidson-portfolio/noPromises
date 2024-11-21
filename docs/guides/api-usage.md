# noPromises API Usage Guide

This guide covers how to interact with the noPromises API to create and manage Flow-Based Programming networks.

## API Endpoints

### Flow Management

#### Creating Flows
```bash
# Create a new flow
curl -X POST http://localhost:8080/api/v1/flows \
    -H "Content-Type: application/json" \
    -d @flow.json
```

#### Managing Flows
```bash
# Start a flow
curl -X POST http://localhost:8080/api/v1/flows/{flow-id}/start

# Stop a flow
curl -X POST http://localhost:8080/api/v1/flows/{flow-id}/stop

# Get flow status
curl http://localhost:8080/api/v1/flows/{flow-id}/status
```

### Network Visualization

Access network visualizations through:
- Static view: `http://localhost:8080/diagrams/network/{flow-id}`
- Live updates: `http://localhost:8080/diagrams/network/{flow-id}/live`

## Flow Configuration

### Basic Flow Structure
```json
{
    "id": "example-flow",
    "nodes": {
        "node-id": {
            "type": "ProcessType",
            "config": {
                // Process-specific configuration
            }
        }
    },
    "edges": [
        {
            "fromNode": "source-node",
            "fromPort": "out",
            "toNode": "target-node",
            "toPort": "in"
        }
    ]
}
```

## Error Handling

The API uses standard HTTP status codes:
- 200: Success
- 400: Bad Request
- 404: Not Found
- 500: Server Error

Error responses include detailed information:
```json
{
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable message",
        "details": {}
    }
}
```

## WebSocket Integration

For real-time updates, connect to:
```javascript
const ws = new WebSocket(`ws://localhost:8080/flows/{flow-id}/events`)
```

Note: Additional API documentation is available through the Swagger UI at `http://localhost:8080/api-docs`
