package middleware

import (
	compress "github.com/GlebRadaev/shlink/internal/middleware/compress"
	http "github.com/GlebRadaev/shlink/internal/middleware/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Middleware applies both base and advanced middleware to the provided router.
func Middleware(r *chi.Mux) {
	AddBaseMiddlewares(r)
	AddAdvancedMiddlewares(r)
}

// AddBaseMiddlewares adds essential middleware that should be available globally across the routes.
func AddBaseMiddlewares(r *chi.Mux) {
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
}

// AddAdvancedMiddlewares applies additional middleware like request logging and compression.
func AddAdvancedMiddlewares(r *chi.Mux) {
	r.Use(http.RequestMiddleware)
	r.Use(compress.CompressMiddleware)
}
