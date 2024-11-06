package api

import (
	"github.com/GlebRadaev/shlink/internal/api/handlers"
	"github.com/go-chi/chi/v5"
)

func Routes(r *chi.Mux, urlHandlers *handlers.URLHandlers, healthHandlers *handlers.HealthHandlers) {
	r.Post("/", urlHandlers.Shorten)
	r.Get("/{id}", urlHandlers.Redirect)
	r.Post("/api/shorten", urlHandlers.ShortenJSON)

	r.Get("/ping", healthHandlers.Ping)
}
