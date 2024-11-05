package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/GlebRadaev/shlink/internal/api"
	"github.com/GlebRadaev/shlink/internal/api/handlers"
	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/middleware"
	"github.com/GlebRadaev/shlink/internal/repository"
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

	log, err := logger.NewLogger("info")
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer log.Sync()

	r := chi.NewRouter()
	middleware.Middleware(r)
	repositories := repository.NewRepositoryFactory(cfg)
	services := service.NewServiceFactory(cfg, log, repositories)
	urlHandlers := handlers.NewURLHandlers(services.URLService)
	api.Routes(r, urlHandlers)

	go func() {
		log.Named("Starting server").Infoln(
			"address", cfg.ServerAddress,
			"baseURL", cfg.BaseURL,
			"fileStoragePath", cfg.FileStoragePath,
		)
		if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
			log.Fatalf("server failed: %v", err)
		}
	}()
	return handleSignals(services.URLService, log)
}

func handleSignals(urlService *service.URLService, log *logger.Logger) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
	log.Error("Received shutdown signal")
	if err := urlService.SaveData(); err != nil {
		log.Errorf("Failed to save data: %v", err)
	} else {
		log.Info("Data successfully saved before shutdown")
	}
	return nil
}
