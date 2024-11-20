# Database Subsystem

## Overview
The noPromises server uses SQLite as its primary database, providing persistent storage with WAL (Write-Ahead Logging) mode for improved concurrency.

## Core Components

### SQLite Database
```go
type SQLiteDB struct {
    db *sql.DB
}

func NewSQLiteDB(path string) (*SQLiteDB, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }

    // Enable WAL mode for better concurrency
    if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
        db.Close()
        return nil, err
    }

    // Enable foreign key constraints
    if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
        db.Close()
        return nil, err
    }

    return &SQLiteDB{db: db}, nil
}
```

## Configuration

### Server Configuration
```go
type Config struct {
    Port     int    // HTTP server port
    DocsPath string // Documentation path
    DBPath   string // SQLite database path (default: "noPromises.db")
}
```

## Best Practices

### Database Operations
- Use prepared statements for queries
- Handle transaction rollbacks
- Implement proper connection pooling
- Close resources appropriately

### Error Handling
- Handle database connection errors
- Implement retry mechanisms
- Log database errors appropriately
- Clean up resources on error

### Testing
- Use temporary databases for tests
- Clean up test databases
- Test WAL mode functionality
- Verify foreign key constraints

## Example Usage

### Database Initialization
```go
func main() {
    db, err := db.NewSQLiteDB("noPromises.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
}
```

### Server Integration
```go
func NewServer(config Config) (*Server, error) {
    if config.DBPath == "" {
        config.DBPath = "noPromises.db"
    }

    db, err := db.NewSQLiteDB(config.DBPath)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize database: %w", err)
    }

    // ... rest of server initialization
}
``` 