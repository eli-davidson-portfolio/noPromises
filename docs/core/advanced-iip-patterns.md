# Advanced IIP Patterns

## Overview

Initial Information Packets (IIPs) provide the primary mechanism for configuring and initializing components in Flow-Based Programming (FBP). This document covers advanced patterns and implementations for IIP usage.

## Core Concepts

### IIP Structure

```go
type IIP[T any] struct {
    Data     T
    Metadata map[string]any
    Port     string
    Priority int  // Higher priority IIPs are delivered first
}

type IIPManager struct {
    iips     map[string][]*IIP[any]  // Map of node ID to IIPs
    delivered map[string]bool        // Track delivery status
    mu       sync.RWMutex
}
```

### IIP Configuration Types

```go
// Strongly typed configuration
type ComponentConfig[T any] struct {
    ConfigType string           // Type identifier for validation
    Data       T               // Typed configuration data
    Schema     *ConfigSchema   // JSON schema for validation
    Version    string          // Configuration version
}

// Dynamic configuration
type DynamicConfig struct {
    ConfigType string
    Data       map[string]interface{}
    Schema     *ConfigSchema
    Version    string
}
```

## Advanced Patterns

### 1. Dynamic Configuration Generation

```go
type ConfigGenerator struct {
    templates map[string]*ConfigTemplate
    resolver  *ConfigResolver
    validator *ConfigValidator
}

func (g *ConfigGenerator) GenerateConfig(ctx context.Context, nodeType string, params map[string]interface{}) (*IIP[any], error) {
    // Get template for node type
    template := g.templates[nodeType]
    if template == nil {
        return nil, fmt.Errorf("no template for node type: %s", nodeType)
    }
    
    // Resolve template variables
    config, err := g.resolver.Resolve(template, params)
    if err != nil {
        return nil, fmt.Errorf("resolving config: %w", err)
    }
    
    // Validate configuration
    if err := g.validator.Validate(config); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return &IIP[any]{
        Data: config,
        Metadata: map[string]any{
            "generated": true,
            "timestamp": time.Now(),
        },
    }, nil
}
```

### 2. IIP Version Management

```go
type VersionedIIP[T any] struct {
    IIP[T]
    Version    string
    Migrations []Migration[T]
}

type Migration[T any] struct {
    FromVersion string
    ToVersion   string
    Migrate     func(T) (T, error)
}

func (v *VersionedIIP[T]) EnsureVersion(targetVersion string) error {
    if v.Version == targetVersion {
        return nil
    }
    
    // Find migration path
    path := v.findMigrationPath(v.Version, targetVersion)
    if path == nil {
        return fmt.Errorf("no migration path from %s to %s", v.Version, targetVersion)
    }
    
    // Apply migrations
    data := v.Data
    var err error
    for _, migration := range path {
        data, err = migration.Migrate(data)
        if err != nil {
            return fmt.Errorf("migration failed: %w", err)
        }
    }
    
    v.Data = data
    v.Version = targetVersion
    return nil
}
```

### 3. Complex Initialization Dependencies

```go
type InitializationManager struct {
    dependencies map[string][]string  // Node ID to dependent node IDs
    initialized  map[string]bool
    mu          sync.RWMutex
}

func (m *InitializationManager) AddDependency(nodeID, dependsOn string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.dependencies[nodeID] = append(m.dependencies[nodeID], dependsOn)
}

func (m *InitializationManager) CanInitialize(nodeID string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    // Check if all dependencies are initialized
    for _, dep := range m.dependencies[nodeID] {
        if !m.initialized[dep] {
            return false
        }
    }
    
    return true
}
```

### 4. Validation Strategies

```go
type IIPValidator struct {
    schemas    map[string]*Schema
    validators map[string]ValidateFunc
}

type ValidateFunc func(interface{}) error

func (v *IIPValidator) ValidateIIP(iip *IIP[any]) error {
    // Get schema for config type
    schema := v.schemas[iip.ConfigType]
    if schema == nil {
        return fmt.Errorf("no schema for config type: %s", iip.ConfigType)
    }
    
    // Validate against schema
    if err := schema.Validate(iip.Data); err != nil {
        return fmt.Errorf("schema validation failed: %w", err)
    }
    
    // Run custom validators
    if validator := v.validators[iip.ConfigType]; validator != nil {
        if err := validator(iip.Data); err != nil {
            return fmt.Errorf("custom validation failed: %w", err)
        }
    }
    
    return nil
}
```

## Implementation Patterns

### 1. Lazy IIP Resolution

```go
type LazyIIP[T any] struct {
    resolver  func() (T, error)
    resolved  T
    hasValue  bool
    mu        sync.RWMutex
}

func (l *LazyIIP[T]) Get() (T, error) {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    if l.hasValue {
        return l.resolved, nil
    }
    
    value, err := l.resolver()
    if err != nil {
        return *new(T), err
    }
    
    l.resolved = value
    l.hasValue = true
    return value, nil
}
```

### 2. IIP Transformation Chain

```go
type IIPTransformer[In, Out any] interface {
    Transform(In) (Out, error)
}

type TransformChain[T any] struct {
    transformers []IIPTransformer[T, T]
}

func (c *TransformChain[T]) Apply(input T) (T, error) {
    current := input
    
    for _, transformer := range c.transformers {
        var err error
        current, err = transformer.Transform(current)
        if err != nil {
            return *new(T), fmt.Errorf("transform failed: %w", err)
        }
    }
    
    return current, nil
}
```

## Best Practices

### 1. Configuration Design
- Use strong typing where possible
- Implement versioning
- Validate configurations
- Provide defaults
- Document schemas

### 2. Initialization Order
- Handle dependencies correctly
- Use deterministic ordering
- Implement timeouts
- Handle circular dependencies
- Log initialization

### 3. Error Handling
- Validate early
- Provide clear errors
- Handle missing configs
- Support defaults
- Log validation failures

### 4. Resource Management
- Clean up on failure
- Handle partial initialization
- Implement timeouts
- Track resource usage
- Monitor initialization

## Common Use Cases

### 1. Database Configuration
```go
type DBConfig struct {
    ConnectionString string
    MaxConnections  int
    Timeout         time.Duration
    RetryPolicy    RetryConfig
}

func NewDatabaseIIP(config DBConfig) *IIP[DBConfig] {
    return &IIP[DBConfig]{
        Data: config,
        Metadata: map[string]any{
            "type": "database",
            "version": "1.0",
        },
    }
}
```

### 2. Feature Flags
```go
type FeatureFlags struct {
    Enabled  map[string]bool
    Variants map[string]string
}

func NewFeatureFlagIIP(flags FeatureFlags) *IIP[FeatureFlags] {
    return &IIP[FeatureFlags]{
        Data: flags,
        Priority: 100, // High priority to configure early
        Metadata: map[string]any{
            "type": "features",
            "updateable": true,
        },
    }
}
```

### 3. Component Chain Configuration
```go
type ChainConfig struct {
    Components []ComponentConfig
    Links      []LinkConfig
}

func NewChainIIP(config ChainConfig) *IIP[ChainConfig] {
    return &IIP[ChainConfig]{
        Data: config,
        Metadata: map[string]any{
            "type": "chain",
            "version": "1.0",
        },
    }
}
```

## Testing Patterns

### 1. Configuration Validation
```go
func TestConfigValidation(t *testing.T) {
    validator := NewIIPValidator()
    
    tests := []struct {
        name    string
        config  interface{}
        wantErr bool
    }{
        {
            name: "valid config",
            config: DBConfig{
                ConnectionString: "valid:conn",
                MaxConnections: 10,
            },
            wantErr: false,
        },
        // Add more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            iip := NewIIP(tt.config)
            err := validator.ValidateIIP(iip)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateIIP() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 2. Version Migration Testing
```go
func TestVersionMigration(t *testing.T) {
    config := OldConfig{Value: "test"}
    iip := NewVersionedIIP(config, "1.0")
    
    err := iip.EnsureVersion("2.0")
    require.NoError(t, err)
    
    newConfig, ok := iip.Data.(NewConfig)
    require.True(t, ok)
    assert.Equal(t, "test", newConfig.NewValue)
}
```

## Implementation Notes

1. Use strong typing whenever possible
2. Implement proper validation
3. Handle version migrations gracefully
4. Support lazy loading for expensive configurations
5. Implement proper cleanup
6. Maintain clear documentation
7. Add comprehensive tests

This document serves as a comprehensive guide to advanced IIP usage in our FBP implementation. For basic IIP concepts, refer to the core documentation.