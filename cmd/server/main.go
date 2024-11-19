package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/elleshadow/noPromises/pkg/server"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Server port")
	docsPath := flag.String("docs", "./docs", "Path to documentation files")
	flag.Parse()

	// Create server
	srv, err := server.NewServer(server.Config{
		Port:     *port,
		DocsPath: *docsPath,
	})
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, initiating shutdown", sig)
		cancel()
	}()

	// Start server
	log.Printf("Starting server on port %d", *port)
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
