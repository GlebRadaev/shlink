package main

import (
	"log"
	"net/http"
	"time"

	"github.com/GlebRadaev/shlink/internal/api"
	"github.com/GlebRadaev/shlink/internal/api/handlers"
	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		if duration > 60*time.Second {
			log.Printf("Request is too slow: %s %s is completed in %v", r.Method, r.URL.Path, duration)
		}
	})
}

func main() {
	cfg := config.ParseAndLoadConfig()
	memoryStorage := repository.NewMemoryStorage()
	urlService := service.NewURLService(memoryStorage)
	urlHandlers := handlers.NewURLHandlers(urlService)

	r := chi.NewRouter()
	r.Use(RequestMiddleware)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	api.RegisterRoutes(r, urlHandlers)

	log.Printf("Server is running on %s", cfg.ServerAddress)
	log.Printf("Base URL is %s", cfg.BaseURL)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
