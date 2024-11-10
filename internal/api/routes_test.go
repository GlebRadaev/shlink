package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/GlebRadaev/shlink/internal/api/handlers"
	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRoutes(t *testing.T) {
	ctx := context.Background()
	cfg, _ := config.ParseAndLoadConfig()
	logger, _ := logger.NewLogger("info")

	repositories := repository.NewRepositoryFactory(ctx, cfg, logger)
	services := service.NewServiceFactory(ctx, cfg, logger, repositories)

	healthHandlers := handlers.NewHealthHandlers(services.HealthService)
	urlHandlers := handlers.NewURLHandlers(services.URLService)

	r := chi.NewRouter()
	Routes(r, urlHandlers, healthHandlers)

	tests := []struct {
		name       string
		method     string
		url        string
		statusCode int
		body       []byte
		headers    map[string]string
	}{
		{
			name:       "Test POST / for shortening",
			method:     http.MethodPost,
			url:        "/",
			statusCode: http.StatusCreated,
			body:       []byte("http://example.com"),
			headers:    nil,
		},
		// {
		// 	name:       "Test GET /{id} for redirect",
		// 	method:     http.MethodGet,
		// 	url:        "/abc12345",      // Используйте существующий shortID
		// 	statusCode: http.StatusFound, // Предположим, что редирект срабатывает
		// 	body:       nil,
		// },

		{
			name:       "Test POST /api/shorten for JSON",
			method:     http.MethodPost,
			url:        "/api/shorten",
			statusCode: http.StatusCreated,
			body:       []byte(`{"url":"http://example.com"}`),
			headers: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name:       "Test GET /ping for health check",
			method:     http.MethodGet,
			url:        "/ping",
			statusCode: http.StatusOK,
			body:       nil,
			headers:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewReader(tt.body))
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}
