package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GlebRadaev/shlink/internal/api"
	"github.com/GlebRadaev/shlink/internal/api/handlers"
	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/middleware"
	"github.com/GlebRadaev/shlink/internal/repository/inmemory"

	"github.com/GlebRadaev/shlink/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	cfg, err := config.ParseAndLoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := logger.NewLogger()
	defer logger.SyncLogger()

	storage := inmemory.NewMemoryStorage()
	urlService := service.NewURLService(storage, cfg)
	urlHandlers := handlers.NewURLHandlers(urlService)

	r := chi.NewRouter()
	middleware.Middleware(r)
	api.Routes(r, urlHandlers)

	log.Named("Starting server").Infoln(
		"address", cfg.ServerAddress,
		"baseURL", cfg.BaseURL,
	)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
}
