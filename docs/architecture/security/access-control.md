# Access Control Architecture

## Core Components

### Authentication System

```go
type AuthService struct {
    providers   map[string]AuthProvider
    sessions    *SessionManager
    tokens      *TokenManager
    audit       *AuditLogger
}

type AuthProvider interface {
    Authenticate(ctx context.Context, credentials interface{}) (*User, error)
    ValidateCredentials(credentials interface{}) error
    GetProviderName() string
}

// Session Management
type SessionManager struct {
    store     SessionStore
    timeout   time.Duration
    cleaner   *SessionCleaner
}

type Session struct {
    ID        string
    UserID    string
    Created   time.Time
    ExpiresAt time.Time
    Metadata  map[string]interface{}
}
```

### Authorization System

```go
type AuthorizationService struct {
    roles       *RoleManager
    permissions *PermissionManager
    policies    *PolicyManager
    cache       *AuthzCache
}

// Role-Based Access Control (RBAC)
type RoleManager struct {
    roles    map[string]*Role
    hierarchy map[string][]string  // Role inheritance
}

type Role struct {
    Name        string
    Permissions []Permission
    Metadata    map[string]interface{}
}

// Permission Management
type PermissionManager struct {
    permissions map[string]*Permission
    cache       *sync.Map
}

type Permission struct {
    Resource string
    Action   string
    Effect   Effect // Allow/Deny
}
```

### Policy Enforcement

```go
type PolicyEnforcer struct {
    policies  *PolicyManager
    evaluator *PolicyEvaluator
    cache     *PolicyCache
}

type Policy struct {
    ID         string
    Effect     Effect
    Subjects   []string    // Users/Roles
    Resources  []string    // Resource patterns
    Actions    []string    // Allowed actions
    Conditions []Condition // Additional checks
}

func (e *PolicyEnforcer) Enforce(ctx context.Context, request *AuthzRequest) (bool, error) {
    // Check cache first
    if result, ok := e.cache.Get(request.Hash()); ok {
        return result.(bool), nil
    }
    
    // Find applicable policies
    policies := e.policies.FindApplicable(request)
    
    // Evaluate policies
    allowed, err := e.evaluator.Evaluate(ctx, policies, request)
    if err != nil {
        return false, fmt.Errorf("evaluating policies: %w", err)
    }
    
    // Cache result
    e.cache.Set(request.Hash(), allowed)
    
    return allowed, nil
}
```

## Access Control Patterns

### Authentication Middleware

```go
func AuthMiddleware(auth *AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract credentials
            creds, err := extractCredentials(r)
            if err != nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            
            // Authenticate
            user, err := auth.Authenticate(r.Context(), creds)
            if err != nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            
            // Create session
            session, err := auth.sessions.Create(user)
            if err != nil {
                http.Error(w, "session creation failed", http.StatusInternalServerError)
                return
            }
            
            // Set session cookie
            http.SetCookie(w, &http.Cookie{
                Name:     "session",
                Value:    session.ID,
                Expires:  session.ExpiresAt,
                HttpOnly: true,
                Secure:   true,
                SameSite: http.SameSiteStrictMode,
            })
            
            // Add user to context
            ctx := context.WithValue(r.Context(), UserKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Authorization Middleware

```go
func AuthzMiddleware(authz *AuthorizationService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Get user from context
            user := GetUserFromContext(r.Context())
            if user == nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            
            // Create authorization request
            request := &AuthzRequest{
                Subject:  user.ID,
                Resource: r.URL.Path,
                Action:   r.Method,
            }
            
            // Check authorization
            allowed, err := authz.IsAllowed(r.Context(), request)
            if err != nil {
                http.Error(w, "authorization check failed", http.StatusInternalServerError)
                return
            }
            
            if !allowed {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Flow Access Control

```go
type FlowAccessController struct {
    authz    *AuthorizationService
    policies *PolicyManager
}

func (c *FlowAccessController) CheckAccess(ctx context.Context, flow *Flow, action string) error {
    user := GetUserFromContext(ctx)
    if user == nil {
        return ErrUnauthorized
    }
    
    request := &AuthzRequest{
        Subject:  user.ID,
        Resource: fmt.Sprintf("flow:%s", flow.ID),
        Action:   action,
    }
    
    allowed, err := c.authz.IsAllowed(ctx, request)
    if err != nil {
        return fmt.Errorf("checking flow access: %w", err)
    }
    
    if !allowed {
        return ErrForbidden
    }
    
    return nil
}
```

## Configuration Management

### Policy Configuration

```go
type PolicyConfig struct {
    Version    string    `json:"version"`
    Policies   []Policy  `json:"policies"`
    Roles      []Role    `json:"roles"`
}

func LoadPolicyConfig(path string) (*PolicyConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading policy config: %w", err)
    }
    
    var config PolicyConfig
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("parsing policy config: %w", err)
    }
    
    return &config, nil
}
```

### Role Configuration

```go
type RoleConfig struct {
    Roles       map[string]Role       `json:"roles"`
    Inheritance map[string][]string   `json:"inheritance"`
}

func (c *RoleConfig) Validate() error {
    // Check for circular dependencies
    if err := validateRoleHierarchy(c.Inheritance); err != nil {
        return fmt.Errorf("invalid role hierarchy: %w", err)
    }
    
    // Validate role definitions
    for name, role := range c.Roles {
        if err := validateRole(name, role); err != nil {
            return fmt.Errorf("invalid role %s: %w", name, err)
        }
    }
    
    return nil
}
```

## Best Practices

### Authentication
- Use secure password hashing
- Implement MFA support
- Rotate session tokens
- Set secure cookie flags
- Monitor auth failures

### Authorization
- Principle of least privilege
- Regular policy review
- Audit access decisions
- Cache auth decisions
- Monitor policy changes

### Session Management
- Secure session storage
- Regular session cleanup
- Session timeout enforcement
- Handle concurrent sessions
- Monitor session activity

### Policy Management
- Version control policies
- Regular policy review
- Policy testing
- Audit policy changes
- Monitor policy enforcement

## Testing

### Authentication Testing

```go
func TestAuthentication(t *testing.T) {
    auth := NewAuthService()
    
    t.Run("valid credentials", func(t *testing.T) {
        creds := Credentials{
            Username: "test",
            Password: "password",
        }
        
        user, err := auth.Authenticate(context.Background(), creds)
        require.NoError(t, err)
        assert.NotNil(t, user)
    })
    
    t.Run("invalid credentials", func(t *testing.T) {
        creds := Credentials{
            Username: "test",
            Password: "wrong",
        }
        
        _, err := auth.Authenticate(context.Background(), creds)
        assert.Error(t, err)
    })
}
```

### Authorization Testing

```go
func TestAuthorization(t *testing.T) {
    authz := NewAuthorizationService()
    
    t.Run("policy evaluation", func(t *testing.T) {
        request := &AuthzRequest{
            Subject:  "user1",
            Resource: "flow:test",
            Action:   "read",
        }
        
        allowed, err := authz.IsAllowed(context.Background(), request)
        require.NoError(t, err)
        assert.True(t, allowed)
    })
}
```

## Implementation Guidelines

1. **Authentication Setup**
   - Configure auth providers
   - Set up session management
   - Implement MFA if needed
   - Configure audit logging

2. **Authorization Setup**
   - Define roles and permissions
   - Configure policies
   - Set up policy enforcement
   - Enable monitoring

3. **Session Management**
   - Configure session storage
   - Set timeouts
   - Enable cleanup
   - Monitor sessions

4. **Policy Management**
   - Version control policies
   - Set up review process
   - Configure monitoring
   - Enable auditing

## Monitoring & Auditing

### Access Monitoring
```go
type AccessMonitor struct {
    metrics  *AccessMetrics
    alerts   *AlertManager
    logger   *AuditLogger
}

func (m *AccessMonitor) RecordAccess(ctx context.Context, event AccessEvent) error {
    // Record metrics
    m.metrics.RecordAccess(event)
    
    // Check for suspicious activity
    if m.isSuspicious(event) {
        m.alerts.Alert(AlertLevelWarning, "Suspicious access detected", event)
    }
    
    // Log access
    return m.logger.LogAccess(ctx, event)
}
```

### Audit Logging
```go
type AuditLogger struct {
    logger  *zap.Logger
    filter  *AuditFilter
}

func (l *AuditLogger) LogAccess(ctx context.Context, event AccessEvent) error {
    // Apply audit filter
    if !l.filter.ShouldLog(event) {
        return nil
    }
    
    // Log event
    l.logger.Info("access event",
        zap.String("user", event.UserID),
        zap.String("resource", event.Resource),
        zap.String("action", event.Action),
        zap.Time("time", event.Time),
        zap.Any("metadata", event.Metadata),
    )
    
    return nil
}
```

## Security Checklist

1. **Authentication Security**
   - [ ] Secure credential storage
   - [ ] MFA configuration
   - [ ] Session security
   - [ ] Token rotation

2. **Authorization Security**
   - [ ] Policy validation
   - [ ] Role hierarchy review
   - [ ] Permission audit
   - [ ] Access monitoring

3. **Session Security**
   - [ ] Secure storage
   - [ ] Timeout configuration
   - [ ] Cleanup procedures
   - [ ] Activity monitoring

4. **Policy Security**
   - [ ] Version control
   - [ ] Change management
   - [ ] Regular review
   - [ ] Audit logging