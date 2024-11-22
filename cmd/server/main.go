package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/elleshadow/noPromises/pkg/server"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Server port")
	docsPath := flag.String("docs", "", "Path to documentation files")
	dbPath := flag.String("db", "noPromises.db", "Path to SQLite database file")
	migrationsPath := flag.String("migrations", "internal/db/migrations", "Path to database migrations")
	flag.Parse()

	// Get absolute path for docs
	absDocsPath := *docsPath
	if absDocsPath == "" {
		// If no docs path provided, use default relative to current directory
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current directory: %v", err)
		}
		absDocsPath = filepath.Join(cwd, "docs")
	}

	// Create and configure server
	srv, err := server.NewServer(server.Config{
		Port:           *port,
		DocsPath:       absDocsPath,
		DBPath:         *dbPath,
		MigrationsPath: *migrationsPath,
	})
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	log.Printf("Server starting on http://localhost:%d", *port)
	log.Printf("Documentation path: %s", absDocsPath)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), srv); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
