package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/elleshadow/noPromises/pkg/server"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	// Create and configure server
	srv, err := server.NewServer(server.Config{
		Port: *port,
	})
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	log.Printf("Server starting on http://localhost:%d", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), srv); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
