package main

import (
	"log"

	"github.com/soloviev1d/url-shortener/server"
)

func main() {
	s, err := server.NewServer(":8080")
	if err != nil {
		log.Fatalf("failed to initialize server: %v\n", err)
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("faile to start the server: %v\n", err)
	}

}
