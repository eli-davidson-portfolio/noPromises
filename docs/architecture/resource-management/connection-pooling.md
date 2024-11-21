# Connection Pool Architecture

## Core Components

### Connection Pool

```go
// Generic connection pool
type Pool[T any] struct {
    active    chan T
    idle      chan T
    factory   func() (T, error)
    validate  func(T) error
    cleanup   func(T) error
    metrics   *PoolMetrics
    mu        sync.RWMutex
}

func NewPool[T any](config PoolConfig[T]) *Pool[T] {
    return &Pool[T]{
        active:    make(chan T, config.MaxActive),
        idle:      make(chan T, config.MaxIdle),
        factory:   config.Factory,
        validate:  config.Validate,
        cleanup:   config.Cleanup,
        metrics:   NewPoolMetrics(),
    }
}

// Acquire connection from pool
func (p *Pool[T]) Acquire(ctx context.Context) (T, error) {
    p.metrics.RecordAcquireAttempt()
    
    // Try to get idle connection
    select {
    case conn := <-p.idle:
        // Validate idle connection
        if err := p.validate(conn); err != nil {
            p.metrics.RecordValidationFailure()
            p.cleanup(conn)
            return p.createConnection()
        }
        p.metrics.RecordIdleHit()
        return conn, nil
    default:
        // No idle connections, try active pool
        select {
        case conn := <-p.active:
            p.metrics.RecordActiveHit()
            return conn, nil
        default:
            // Create new connection
            return p.createConnection()
        }
    }
}
```

### Connection Management

```go
// Connection wrapper with metadata
type ManagedConnection[T any] struct {
    Conn      T
    Created   time.Time
    LastUsed  time.Time
    UseCount  int64
    Errors    int64
}

func (c *ManagedConnection[T]) MarkUsed() {
    c.LastUsed = time.Now()
    atomic.AddInt64(&c.UseCount, 1)
}

func (c *ManagedConnection[T]) MarkError() {
    atomic.AddInt64(&c.Errors, 1)
}

// Pool maintenance
func (p *Pool[T]) Maintain(ctx context.Context) {
    ticker := time.NewTicker(p.config.MaintenanceInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            p.removeStaleConnections()
            p.ensureMinConnections()
            p.balancePoolSize()
        }
    }
}
```

### Health Checking

```go
type HealthChecker[T any] struct {
    validate      func(T) error
    maxErrors     int64
    checkInterval time.Duration
    metrics       *HealthMetrics
}

func (h *HealthChecker[T]) CheckConnection(conn *ManagedConnection[T]) error {
    // Check error threshold
    if conn.Errors >= h.maxErrors {
        h.metrics.RecordErrorThresholdExceeded()
        return ErrTooManyErrors
    }
    
    // Check age
    if time.Since(conn.Created) > h.config.MaxAge {
        h.metrics.RecordAgeExceeded()
        return ErrTooOld
    }
    
    // Validate connection
    if err := h.validate(conn.Conn); err != nil {
        conn.MarkError()
        h.metrics.RecordValidationFailure()
        return err
    }
    
    return nil
}
```

## Advanced Features

### Connection Events

```go
type PoolEvent struct {
    Type      EventType
    ConnID    string
    Timestamp time.Time
    Metadata  map[string]interface{}
}

type PoolEventHandler interface {
    OnConnectionCreated(*PoolEvent)
    OnConnectionClosed(*PoolEvent)
    OnConnectionError(*PoolEvent)
    OnPoolExhausted(*PoolEvent)
}

func (p *Pool[T]) notifyEvent(event *PoolEvent) {
    for _, handler := range p.eventHandlers {
        go handler.HandleEvent(event)
    }
}
```

### Adaptive Sizing

```go
type AdaptivePool[T any] struct {
    *Pool[T]
    minSize     int
    maxSize     int
    loadFactor  float64
}

func (p *AdaptivePool[T]) adjustSize() {
    metrics := p.metrics.GetRecentMetrics()
    
    // Calculate optimal size based on usage patterns
    utilization := float64(metrics.ActiveConnections) / float64(p.currentSize())
    
    if utilization > p.loadFactor {
        // Grow pool
        p.grow(p.calculateGrowth(utilization))
    } else if utilization < p.loadFactor/2 {
        // Shrink pool
        p.shrink(p.calculateShrinkage(utilization))
    }
}

func (p *AdaptivePool[T]) calculateGrowth(utilization float64) int {
    currentSize := p.currentSize()
    growthFactor := (utilization - p.loadFactor) + 1.0
    newSize := int(float64(currentSize) * growthFactor)
    return min(newSize, p.maxSize)
}
```

### Background Cleanup

```go
type PoolCleaner[T any] struct {
    pool         *Pool[T]
    maxIdleTime  time.Duration
    maxAge       time.Duration
}

func (c *PoolCleaner[T]) Clean(ctx context.Context) {
    ticker := time.NewTicker(c.config.CleanupInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            c.cleanIdleConnections()
            c.cleanOldConnections()
        }
    }
}

func (c *PoolCleaner[T]) cleanIdleConnections() {
    threshold := time.Now().Add(-c.maxIdleTime)
    
    for {
        select {
        case conn := <-c.pool.idle:
            if conn.LastUsed.Before(threshold) {
                c.pool.metrics.RecordIdleTimeout()
                c.pool.cleanup(conn.Conn)
            } else {
                c.pool.idle <- conn
                return
            }
        default:
            return
        }
    }
}
```

## Resource Management

### Connection Lifecycle

```go
type ConnectionManager[T any] struct {
    pool      *Pool[T]
    tracker   *ResourceTracker
    recycler  *ConnectionRecycler[T]
}

func (m *ConnectionManager[T]) CreateConnection() (T, error) {
    conn, err := m.pool.factory()
    if err != nil {
        return *new(T), err
    }
    
    managed := &ManagedConnection[T]{
        Conn:    conn,
        Created: time.Now(),
    }
    
    m.tracker.Track(managed)
    return conn, nil
}

func (m *ConnectionManager[T]) RecycleConnection(conn T) {
    if managed := m.tracker.Get(conn); managed != nil {
        m.recycler.Recycle(managed)
    }
}
```

### Resource Tracking

```go
type ResourceTracker struct {
    resources map[interface{}]*ResourceInfo
    mu        sync.RWMutex
}

type ResourceInfo struct {
    Created   time.Time
    LastUsed  time.Time
    UseCount  int64
    Metadata  map[string]interface{}
}

func (t *ResourceTracker) Track(resource interface{}) {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    t.resources[resource] = &ResourceInfo{
        Created:  time.Now(),
        Metadata: make(map[string]interface{}),
    }
}
```

## Monitoring & Metrics

### Pool Metrics

```go
type PoolMetrics struct {
    // Gauges
    ActiveConnections  prometheus.Gauge
    IdleConnections    prometheus.Gauge
    
    // Counters
    ConnectionsCreated prometheus.Counter
    ConnectionsClosed  prometheus.Counter
    
    // Histograms
    AcquisitionTime   prometheus.Histogram
    ConnectionAge     prometheus.Histogram
    
    // Health metrics
    ValidationFailures prometheus.Counter
    ErrorRates        *prometheus.CounterVec
}

func (m *PoolMetrics) RecordAcquisition(duration time.Duration) {
    m.AcquisitionTime.Observe(duration.Seconds())
}

func (m *PoolMetrics) RecordConnectionError(errorType string) {
    m.ErrorRates.WithLabelValues(errorType).Inc()
}
```

### Health Monitoring

```go
type HealthMonitor[T any] struct {
    pool      *Pool[T]
    checker   *HealthChecker[T]
    alerts    *AlertManager
}

func (m *HealthMonitor[T]) Monitor(ctx context.Context) {
    ticker := time.NewTicker(m.config.CheckInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.checkPoolHealth()
        }
    }
}

func (m *HealthMonitor[T]) checkPoolHealth() {
    metrics := m.pool.metrics.GetRecentMetrics()
    
    // Check error rates
    if metrics.ErrorRate() > m.config.MaxErrorRate {
        m.alerts.Alert(AlertLevelWarn, "High connection error rate detected")
    }
    
    // Check pool exhaustion
    if metrics.PoolExhaustion() > m.config.MaxExhaustionRate {
        m.alerts.Alert(AlertLevelError, "Pool exhaustion detected")
    }
}
```

## Best Practices

### Pool Configuration
- Set appropriate pool sizes
- Configure timeouts
- Enable health checks
- Monitor pool metrics
- Configure cleanup

### Resource Management
- Track connection lifecycle
- Implement proper cleanup
- Handle connection errors
- Monitor resource usage
- Configure alerts

### Error Handling
- Handle connection failures
- Implement retry logic
- Set error thresholds
- Clean up on errors
- Log error patterns

### Performance
- Monitor acquisition times
- Track pool utilization
- Optimize pool size
- Handle backpressure
- Monitor metrics

## Testing

### Pool Testing
```go
func TestConnectionPool(t *testing.T) {
    pool := NewPool(PoolConfig{
        MaxActive: 10,
        MaxIdle:   5,
        Factory:   createTestConnection,
        Validate:  validateTestConnection,
    })
    
    t.Run("acquisition", func(t *testing.T) {
        conn, err := pool.Acquire(context.Background())
        require.NoError(t, err)
        assert.NotNil(t, conn)
        
        pool.Release(conn)
    })
    
    t.Run("concurrent usage", func(t *testing.T) {
        var wg sync.WaitGroup
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                conn, err := pool.Acquire(context.Background())
                require.NoError(t, err)
                time.Sleep(time.Millisecond)
                pool.Release(conn)
            }()
        }
        wg.Wait()
    })
}
```

### Health Check Testing
```go
func TestHealthChecks(t *testing.T) {
    checker := NewHealthChecker(HealthConfig{
        MaxErrors: 3,
        CheckInterval: time.Second,
    })
    
    t.Run("error threshold", func(t *testing.T) {
        conn := &ManagedConnection{
            Errors: 4,
        }
        
        err := checker.CheckConnection(conn)
        assert.Error(t, err)
        assert.Equal(t, ErrTooManyErrors, err)
    })
}
```

## Implementation Checklist

1. **Basic Setup**
   - [ ] Configure pool sizes
   - [ ] Implement factory
   - [ ] Set up validation
   - [ ] Configure cleanup

2. **Health Management**
   - [ ] Enable health checks
   - [ ] Configure monitoring
   - [ ] Set up alerts
   - [ ] Implement cleanup

3. **Resource Tracking**
   - [ ] Track connections
   - [ ] Monitor usage
   - [ ] Configure logging
   - [ ] Set up metrics

4. **Testing**
   - [ ] Unit tests
   - [ ] Load tests
   - [ ] Error scenarios
   - [ ] Performance tests