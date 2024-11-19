# API JSON Schemas

## Overview
This document defines the JSON schemas for all request and response objects in the noPromises API. These schemas are used for validation and documentation.

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
            "pattern": "^[a-zA-Z0-9-_]+$",
            "minLength": 1,
            "maxLength": 64,
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
            },
            "minItems": 1
        },
        "metadata": {
            "type": "object",
            "additionalProperties": true
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

### Edge Configuration
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "id": "#/definitions/EdgeConfig",
    "type": "object",
    "required": ["from", "to"],
    "properties": {
        "from": {
            "$ref": "#/definitions/NodePort"
        },
        "to": {
            "$ref": "#/definitions/NodePort"
        }
    }
}
```

### Node Port Reference
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "id": "#/definitions/NodePort",
    "type": "object",
    "required": ["node", "port"],
    "properties": {
        "node": {
            "type": "string",
            "description": "Node identifier"
        },
        "port": {
            "type": "string",
            "description": "Port name on the node"
        }
    }
}
```

## Status Response Schemas

### Flow Status Response
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "required": ["id", "state", "nodes"],
    "properties": {
        "id": {
            "type": "string"
        },
        "state": {
            "type": "string",
            "enum": ["created", "starting", "running", "stopping", "stopped", "error"]
        },
        "started_at": {
            "type": "string",
            "format": "date-time"
        },
        "error": {
            "type": "string"
        },
        "nodes": {
            "type": "object",
            "additionalProperties": {
                "$ref": "#/definitions/NodeStatus"
            }
        }
    }
}
```

### Node Status
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "id": "#/definitions/NodeStatus",
    "type": "object",
    "required": ["state", "messages_processed", "last_active"],
    "properties": {
        "state": {
            "type": "string",
            "enum": ["created", "starting", "running", "stopping", "stopped", "error"]
        },
        "messages_processed": {
            "type": "integer",
            "minimum": 0
        },
        "last_active": {
            "type": "string",
            "format": "date-time"
        },
        "error": {
            "type": "string"
        }
    }
}
```

## Process Discovery Schemas

### Process Type List Response
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "required": ["processes"],
    "properties": {
        "processes": {
            "type": "array",
            "items": {
                "$ref": "#/definitions/ProcessType"
            }
        }
    }
}
```

### Process Type Definition
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "id": "#/definitions/ProcessType",
    "type": "object",
    "required": ["name", "description", "input_ports", "output_ports", "config"],
    "properties": {
        "name": {
            "type": "string"
        },
        "description": {
            "type": "string"
        },
        "input_ports": {
            "type": "array",
            "items": {
                "$ref": "#/definitions/PortInfo"
            }
        },
        "output_ports": {
            "type": "array",
            "items": {
                "$ref": "#/definitions/PortInfo"
            }
        },
        "config": {
            "$ref": "#/definitions/ConfigSchema"
        }
    }
}
```

### Port Information
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "id": "#/definitions/PortInfo",
    "type": "object",
    "required": ["name", "description", "required"],
    "properties": {
        "name": {
            "type": "string"
        },
        "description": {
            "type": "string"
        },
        "required": {
            "type": "boolean"
        }
    }
}
```

### Configuration Schema
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "id": "#/definitions/ConfigSchema",
    "type": "object",
    "required": ["properties"],
    "properties": {
        "properties": {
            "type": "object",
            "additionalProperties": {
                "$ref": "#/definitions/PropertySchema"
            }
        },
        "required": {
            "type": "array",
            "items": {
                "type": "string"
            }
        }
    }
}
```

### Property Schema
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "id": "#/definitions/PropertySchema",
    "type": "object",
    "required": ["type", "description"],
    "properties": {
        "type": {
            "type": "string",
            "enum": ["string", "number", "integer", "boolean", "object", "array"]
        },
        "description": {
            "type": "string"
        },
        "default": {
            "description": "Default value for the property"
        }
    }
}
```

## Error Response Schema
```json
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "required": ["error"],
    "properties": {
        "error": {
            "type": "object",
            "required": ["code", "message"],
            "properties": {
                "code": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                },
                "details": {
                    "type": "object",
                    "additionalProperties": true
                }
            }
        }
    }
}
```

## Schema Validation
These schemas can be used with standard JSON Schema validators. Example validation using Go:

```go
type Validator struct {
    schemas map[string]*jsonschema.Schema
}

func NewValidator() (*Validator, error) {
    compiler := jsonschema.NewCompiler()
    
    // Load all schemas
    for name, schema := range defaultSchemas {
        if err := compiler.AddResource(name, strings.NewReader(schema)); err != nil {
            return nil, err
        }
    }
    
    return &Validator{
        schemas: compiler.Compile(),
    }, nil
}

func (v *Validator) ValidateFlowConfig(config interface{}) error {
    return v.schemas["flow-config"].Validate(config)
}
```