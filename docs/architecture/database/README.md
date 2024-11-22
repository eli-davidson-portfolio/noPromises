# Database Architecture

## Overview
The database system uses SQLite with WAL mode for improved concurrency and reliability. It includes a robust migrations system for schema management.

## Core Components

### Database Connection
```go
type DB struct {
    *sql.DB
}
```
- Wraps standard `sql.DB`
- Enables WAL mode
- Enforces foreign key constraints
- Provides connection management

### Migration System
```go
type Migration struct {
    Version    int
    Name       string
    UpScript   string
    DownScript string
}
```

The migration system provides:
- Version tracking
- Forward and rollback migrations
- Transaction-safe application
- Schema version management

### Migration Manager
```go
type MigrationManager struct {
    db *DB
}
```
Handles:
- Migration application
- Rollback operations
- Version tracking
- Error handling

## Database Operations

### Connection Setup
```go
db, err := New(dsn)
```
1. Opens SQLite connection
2. Enables WAL mode
3. Enables foreign key constraints
4. Returns wrapped connection

### Migration Management
```go
manager := NewMigrationManager(db)
err := manager.ApplyMigrations(ctx, migrationsDir)
```

### Schema Management
- Migrations stored in SQL files
- Naming format: `{version}_{name}.{up|down}.sql`
- Transaction-safe execution
- Version tracking in `schema_version` table

## Best Practices

1. **Migration Files**
   - Use descriptive names
   - Include both up and down scripts
   - Keep migrations atomic
   - Test both directions

2. **Error Handling**
   - Transaction rollback on errors
   - Proper error wrapping
   - Detailed error messages
   - Clean resource management

3. **Version Management**
   - Sequential version numbers
   - No version gaps
   - No version conflicts
   - Version validation

4. **Testing**
   - Test all migrations
   - Test rollbacks
   - Use in-memory database
   - Verify constraints 