# Resource Lifecycle Management

This document details the resource management system in our FBP implementation, focusing on lifecycle management of system resources.

## Core Components

### Resource Interface

```go
type Resource interface {
    // Initialize sets up the resource
    Initialize(ctx context.Context) error
    
    // Close cleans up the resource
    Close(ctx context.Context) error
    
    // HealthCheck verifies resource is operational
    HealthCheck(ctx context.Context) error
}
```

### Resource Pool

```go
type ResourcePool[T Resource] struct {
    resources chan T
    factory   func() (T, error)
    size      int
    mu        sync.RWMutex
    active    map[T]time.Time
}
```

### Resource Manager

```go
type ResourceManager struct {
    pools    map[string]interface{}
    mu       sync.RWMutex
    ctx      context.Context
    cancel   context.CancelFunc
}
```

## Lifecycle Phases

### 1. Initialization

```go
func (p *ResourcePool[T]) Initialize(ctx context.Context) error {
    for i := 0; i < p.size; i++ {
        resource, err := p.factory()
        if err != nil {
            return fmt.Errorf("initializing resource: %w", err)
        }
        
        if err := resource.Initialize(ctx); err != nil {
            return fmt.Errorf("initializing resource: %w", err)
        }
        
        p.resources <- resource
    }
    return nil
}
```

### 2. Resource Acquisition

```go
func (p *ResourcePool[T]) Acquire(ctx context.Context) (T, error) {
    select {
    case resource := <-p.resources:
        if err := resource.HealthCheck(ctx); err != nil {
            // Resource is unhealthy, create new one
            newResource, err := p.factory()
            if err != nil {
                return resource, fmt.Errorf("creating new resource: %w", err)
            }
            resource = newResource
        }
        
        p.mu.Lock()
        p.active[resource] = time.Now()
        p.mu.Unlock()
        
        return resource, nil
    case <-ctx.Done():
        var zero T
        return zero, ctx.Err()
    }
}
```

### 3. Resource Release

```go
func (p *ResourcePool[T]) Release(resource T) {
    p.mu.Lock()
    delete(p.active, resource)
    p.mu.Unlock()
    
    p.resources <- resource
}
```

## Resource-Aware Components

### Resource-Aware Node

```go
type ResourceAwareNode[In, Out any, R Resource] struct {
    Node[In, Out]
    pool      *ResourcePool[R]
    processor func(ctx context.Context, resource R, data In) (Out, error)
}

func (n *ResourceAwareNode[In, Out, R]) Process(ctx context.Context, in Port[In], out Port[Out]) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
            
        case data, ok := <-in:
            if !ok {
                return nil
            }
            
            // Acquire resource
            resource, err := n.pool.Acquire(ctx)
            if err != nil {
                return fmt.Errorf("acquiring resource: %w", err)
            }
            
            // Process with resource
            result, err := n.processor(ctx, resource, data)
            
            // Release resource
            n.pool.Release(resource)
            
            if err != nil {
                return fmt.Errorf("processing with resource: %w", err)
            }
            
            // Send result
            select {
            case out <- result:
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }
}
```

## Best Practices

1. **Resource Initialization**
   - Initialize resources lazily
   - Validate resource health after creation
   - Handle initialization failures gracefully

2. **Resource Pooling**
   - Size pools appropriately
   - Monitor pool utilization
   - Implement backpressure when needed

3. **Health Checks**
   - Regular health validation
   - Quick failure detection
   - Automatic resource recreation

4. **Cleanup**
   - Proper resource release
   - Complete shutdown procedure
   - Handle partial failures

## Example Usage

### Database Connection Pool
```go
type DBConnection struct {
    conn interface{} // Replace with actual DB connection type
}

func (db *DBConnection) Initialize(ctx context.Context) error {
    // Initialize DB connection
    return nil
}

func (db *DBConnection) Close(ctx context.Context) error {
    // Close DB connection
    return nil
}

func (db *DBConnection) HealthCheck(ctx context.Context) error {
    // Check if connection is alive
    return nil
}

func Example_DatabasePool() {
    // Create pool
    dbPool := NewResourcePool[*DBConnection](10, func() (*DBConnection, error) {
        return &DBConnection{}, nil
    })
    
    // Create resource-aware node
    dbNode := NewResourceAwareNode[string, string, *DBConnection](
        &BaseNode{},
        dbPool,
        func(ctx context.Context, db *DBConnection, query string) (string, error) {
            // Use DB connection to process query
            return "result", nil
        },
    )
    
    // Use in network...
}
```