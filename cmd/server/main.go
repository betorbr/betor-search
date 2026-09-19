package main

import (
	"log"
	"net/http"
	"os"

	"betor-search/internal/health/application"
	"betor-search/internal/health/transport"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/health", transport.NewHandler(application.NewService()))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
