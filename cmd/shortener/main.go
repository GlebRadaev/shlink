package main

import (
	"log"
	"net/http"

	"github.com/GlebRadaev/shlink/internal/api"
	"github.com/GlebRadaev/shlink/internal/api/handlers"
	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/middleware"
	"github.com/GlebRadaev/shlink/internal/repository"

	"github.com/GlebRadaev/shlink/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.ParseAndLoadConfig()
	if err != nil {
		return err
	}
	memoryStorage := repository.NewMemoryStorage()
	urlService := service.NewURLService(memoryStorage, cfg)
	urlHandlers := handlers.NewURLHandlers(urlService)

	r := chi.NewRouter()
	middleware.Middleware(r)
	api.Routes(r, urlHandlers)

	log.Printf("Server is running on %s", cfg.ServerAddress)
	log.Printf("Base URL is %s", cfg.BaseURL)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
