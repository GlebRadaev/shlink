// Package api provides the API routes for the URL shortening service and health checking.
// It sets up routes for handling URL shortening requests, user-specific URL management,
// and health check endpoints. The package uses the `chi` router to handle HTTP requests.
//
// Routes:
// - POST /: Shortens a URL using the URLHandlers.Shorten handler.
// - GET /{id}: Redirects to the original URL based on the provided ID using the URLHandlers.Redirect handler.
// - POST /api/shorten: Shortens a URL based on the JSON body using the URLHandlers.ShortenJSON handler.
// - POST /api/shorten/batch: Shortens multiple URLs in batch using the URLHandlers.ShortenJSONBatch handler.
// - GET /api/user/urls: Fetches all URLs associated with a user using the URLHandlers.GetUserURLs handler.
// - DELETE /api/user/urls: Deletes all URLs associated with a user using the URLHandlers.DeleteUserURLs handler.
// - GET /ping: Returns a health check status using the HealthHandlers.Ping handler.
package api

import (
	"github.com/GlebRadaev/shlink/internal/api/handlers"
	"github.com/go-chi/chi/v5"
)

// Routes sets up API routes for URL shortening and health checking.
func Routes(r *chi.Mux, urlHandlers *handlers.URLHandlers, healthHandlers *handlers.HealthHandlers) {
	r.Post("/", urlHandlers.Shorten)
	r.Get("/{id}", urlHandlers.Redirect)
	r.Post("/api/shorten", urlHandlers.ShortenJSON)
	r.Post("/api/shorten/batch", urlHandlers.ShortenJSONBatch)
	r.Get("/api/user/urls", urlHandlers.GetUserURLs)
	r.Delete("/api/user/urls", urlHandlers.DeleteUserURLs)

	r.Get("/ping", healthHandlers.Ping)
}
