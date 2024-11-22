package db

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	// Create temporary database file
	f, err := os.CreateTemp("", "test-db-*.db")
	require.NoError(t, err)
	f.Close()
	defer os.Remove(f.Name())

	t.Run("new database", func(t *testing.T) {
		db, err := New(f.Name())
		require.NoError(t, err)
		defer db.Close()

		// Test WAL mode
		var journalMode string
		err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
		require.NoError(t, err)
		assert.Equal(t, "wal", journalMode)

		// Test foreign keys
		var foreignKeys int
		err = db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
		require.NoError(t, err)
		assert.Equal(t, 1, foreignKeys)
	})

	t.Run("invalid dsn", func(t *testing.T) {
		db, err := New("/invalid/path/db.sqlite")
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}
