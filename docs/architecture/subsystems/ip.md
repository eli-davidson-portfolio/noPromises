# Information Packet (IP) Subsystem

The IP subsystem provides the fundamental data unit for our FBP implementation.

## Core Components

### IP Structure

```go
type IP[T any] struct {
    Data     T
    Metadata map[string]any
}
```

### Features

#### Generic Type Support
- Type-safe data handling
- Compile-time type checking
- Support for any data type

#### Metadata Management
- Key-value metadata storage
- Creation timestamp tracking
- Owner tracking
- Custom metadata support

### IP Creation

```go
// Create normal IP
packet := ip.New[string]("data")

// Create with metadata
packet := ip.New[int](42)
packet.SetMetadata("key", "value")
```

### Operations

#### Data Access
```go
// Get data
data := packet.Data()

// Get metadata
value, exists := packet.GetMetadata("key")
```

#### Ownership
```go
// Set owner
err := packet.SetOwner("process1")

// Get owner
owner := packet.Owner()
```

## Best Practices

### Type Safety
- Use explicit type parameters
- Handle type conversions carefully
- Validate data types at boundaries

### Metadata Usage
- Use consistent metadata keys
- Document metadata meanings
- Clean up metadata when done

### Error Handling
- Check ownership operations
- Validate metadata keys
- Handle missing metadata gracefully

## Testing

### Test Cases
- Data type handling
- Metadata operations
- Ownership changes
- Creation timestamps
- Error conditions 