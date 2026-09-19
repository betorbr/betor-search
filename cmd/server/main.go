package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"betor-search/internal/health/application"
	"betor-search/internal/health/transport"
	searchApplication "betor-search/internal/search/application"
	searchTransport "betor-search/internal/search/transport"
)

func main() {
	searchInterval := 30 * time.Minute
	if value := os.Getenv("BETOR_SEARCH_UPDATE_INTERVAL_MINUTES"); value != "" {
		if parsedMinutes, err := time.ParseDuration(value + "m"); err == nil {
			searchInterval = parsedMinutes
		}
	}
	catalogService := searchApplication.NewService("", searchInterval)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go catalogService.StartBackgroundSync(ctx)

	healthService := application.NewServiceWithCatalog("betor-search-catalog", catalogService)
	mux := http.NewServeMux()
	mux.Handle("/health", transport.NewHandler(healthService))
	mux.Handle("/v1/search/", http.Handler(searchTransport.NewHandler(catalogService)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
