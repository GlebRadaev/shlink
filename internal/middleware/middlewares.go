package middleware

import (
	http "github.com/GlebRadaev/shlink/internal/middleware/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Middleware(r *chi.Mux) {
	AddBaseMiddlewares(r)
	AddAdvancedMiddlewares(r)
}

func AddBaseMiddlewares(r *chi.Mux) {
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
}

func AddAdvancedMiddlewares(r *chi.Mux) {
	r.Use(http.RequestMiddleware)
}
