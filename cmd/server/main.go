package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/elleshadow/noPromises/pkg/server/docs"
)

func main() {
	// Parse command line flags
	docsPath := flag.String("docs", "./docs", "Path to documentation files")
	port := flag.String("port", "8080", "Server port")
	flag.Parse()

	// Create and configure docs server
	docsServer := docs.NewServer(docs.Config{
		DocsPath: *docsPath,
	})

	// Setup routes
	docsServer.SetupRoutes()

	// Start server
	log.Printf("Documentation server starting on http://localhost:%s/docs", *port)
	if err := http.ListenAndServe(":"+*port, docsServer.Router()); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
