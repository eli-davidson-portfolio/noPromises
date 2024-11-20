package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteDB(t *testing.T) {
	// Create temporary test database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create new database
	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Clean up
	defer func() {
		db.Close()
		os.Remove(dbPath)
		os.Remove(dbPath + "-shm") // Remove WAL files
		os.Remove(dbPath + "-wal")
	}()

	// Test database operations
	t.Run("database operations", func(t *testing.T) {
		// Test WAL mode
		var journalMode string
		err = db.DB().QueryRow("PRAGMA journal_mode").Scan(&journalMode)
		require.NoError(t, err)
		assert.Equal(t, "wal", journalMode)

		// Test foreign keys
		var foreignKeys int
		err = db.DB().QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
		require.NoError(t, err)
		assert.Equal(t, 1, foreignKeys)
	})
}

func TestSQLiteDBErrors(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		db, err := NewSQLiteDB("/invalid/path/test.db")
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}
