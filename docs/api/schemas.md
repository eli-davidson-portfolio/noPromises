# API JSON Schemas

## Flow Configuration Schemas

### Flow Creation Request
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "required": ["id", "nodes", "edges"],
    "properties": {
        "id": {
            "type": "string",
            "description": "Unique identifier for the flow"
        },
        "nodes": {
            "type": "object",
            "additionalProperties": {
                "$ref": "#/definitions/NodeConfig"
            },
            "minProperties": 1
        },
        "edges": {
            "type": "array",
            "items": {
                "$ref": "#/definitions/EdgeConfig"
            }
        }
    }
}
```

### Node Configuration
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "id": "#/definitions/NodeConfig",
    "type": "object",
    "required": ["type", "config"],
    "properties": {
        "type": {
            "type": "string",
            "description": "Process type identifier"
        },
        "config": {
            "type": "object",
            "additionalProperties": true,
            "description": "Process-specific configuration"
        }
    }
}
```

### Flow Response
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "required": ["data"],
    "properties": {
        "data": {
            "type": "object",
            "required": ["id", "config", "state"],
            "properties": {
                "id": {
                    "type": "string"
                },
                "config": {
                    "type": "object"
                },
                "state": {
                    "type": "string",
                    "enum": ["created", "starting", "running", "stopping", "stopped", "error"]
                },
                "started_at": {
                    "type": ["string", "null"],
                    "format": "date-time"
                },
                "error": {
                    "type": "string"
                }
            }
        }
    }
}
```

### Error Response
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "required": ["error"],
    "properties": {
        "error": {
            "type": "object",
            "required": ["message"],
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        }
    }
}
```