# Database Migrations

## Migration Structure

### File Format
```
migrations/
├── 000001_initial_schema.up.sql
├── 000001_initial_schema.down.sql
└── loader.go
```

### Migration Files
- Up migrations: `{version}_{name}.up.sql`
- Down migrations: `{version}_{name}.down.sql`
- Version format: 6-digit number
- Descriptive names

## Migration Management

### Loading Migrations
```go
migrations, err := LoadMigrations(dir)
```
- Reads migration files
- Validates format
- Sorts by version
- Pairs up/down scripts

### Applying Migrations
```go
err := migrator.ApplyMigrations(ctx, migrations)
```
1. Creates version table if needed
2. Checks current version
3. Applies new migrations
4. Records successful migrations

### Rolling Back
```go
err := migrator.RollbackMigration(ctx)
```
1. Gets last applied migration
2. Executes down script
3. Updates version table
4. Handles errors

## Schema Version Table

```sql
CREATE TABLE schema_version (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    up_script TEXT NOT NULL,
    down_script TEXT NOT NULL,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
```

## Best Practices

### Migration Design
1. Make migrations atomic
2. Include rollback logic
3. Test both directions
4. Use transactions

### Version Management
1. Sequential versions
2. No gaps in versions
3. No version conflicts
4. Version validation

### Error Handling
1. Transaction rollback
2. Proper error wrapping
3. Resource cleanup
4. State consistency

### Testing
1. Test all migrations
2. Test rollbacks
3. Use in-memory database
4. Verify constraints 