package main

import (
	"context"
	"log"

	"github.com/elleshadow/noPromises/pkg/server"
)

func main() {
	srv, err := server.NewServer(server.Config{
		Port: 8080,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := srv.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}
