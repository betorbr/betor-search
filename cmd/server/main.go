package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"betor-search/internal/health/application"
	"betor-search/internal/health/transport"
	searchApplication "betor-search/internal/search/application"
	searchTransport "betor-search/internal/search/transport"
)

func main() {
	searchInterval := searchApplication.ParseUpdateIntervalFromEnv()
	if value := strings.TrimSpace(os.Getenv("BETOR_SEARCH_UPDATE_INTERVAL_MINUTES")); value != "" {
		log.Printf("loaded sync interval from env: BETOR_SEARCH_UPDATE_INTERVAL_MINUTES=%s parsed=%s", value, searchInterval)
	}
	log.Printf("starting catalog sync loop with interval=%s", searchInterval)
	catalogService := searchApplication.NewService("", searchInterval)
	if err := catalogService.Sync(); err != nil {
		log.Printf("initial catalog sync failed: %v", err)
	} else {
		log.Printf("initial catalog sync succeeded: last_success=%s status=%s", time.Now().UTC().Format(time.RFC3339Nano), catalogService.Status())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go catalogService.StartBackgroundSync(ctx)

	healthService := application.NewServiceWithCatalog("betor-search-catalog", &catalogService)
	mux := http.NewServeMux()
	mux.Handle("/health", transport.NewHandler(healthService))
	mux.Handle("/v1/search/", http.Handler(searchTransport.NewHandler(catalogService)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
