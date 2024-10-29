package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service"

	"github.com/GlebRadaev/shlink/internal/service/url"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func setup() (*url.URLService, *config.Config, error) {
	cfg, err := config.ParseAndLoadConfig()
	if err != nil {
		return nil, nil, err
	}
	log, _ := logger.NewLogger("info")
	repositories := repository.NewRepositoryFactory(cfg)
	services := service.NewServiceFactory(cfg, log, repositories)
	return services.URLService, cfg, nil
}

func TestURLHandlers_Shorten(t *testing.T) {
	type args struct {
		contentType string
		body        string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantBody   string
	}{
		{
			name: "valid URL",
			args: args{
				contentType: "text/plain",
				body:        "http://example.com",
			},
			wantStatus: http.StatusCreated,
			wantBody:   "http://localhost/shortID",
		},
		{
			name: "invalid URL format (URL too long)",
			args: args{
				contentType: "text/plain",
				body:        strings.Repeat("a", 2048+1),
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid URL format\n",
		},
		{
			name: "invalid URL format",
			args: args{
				contentType: "text/plain",
				body:        "invalid-url",
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid URL format\n",
		},
	}

	urlService, cfg, err := setup()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	handler := NewURLHandlers(urlService)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.args.body))
			req.Header.Set("Content-Type", tt.args.contentType)
			w := httptest.NewRecorder()
			router := chi.NewRouter()
			router.Post("/", handler.Shorten)

			router.ServeHTTP(w, req)

			res := w.Result()
			body, _ := io.ReadAll(res.Body)
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantStatus == http.StatusCreated {
				assert.Contains(t, string(body), cfg.BaseURL+"/")
			} else {
				assert.Equal(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestURLHandlers_Redirect(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name       string
		args       args
		setup      func(service *url.URLService) string
		wantStatus int
	}{
		{
			name: "valid ID",
			args: args{id: "validIDD"},
			setup: func(service *url.URLService) string {
				url, _ := service.Shorten("http://example.com")
				splitURL := strings.Split(url, "/")
				shortID := splitURL[len(splitURL)-1]
				return shortID
			},
			wantStatus: http.StatusTemporaryRedirect,
		},
		{
			name:       "ID not found",
			args:       args{id: "notFoundID"},
			setup:      nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid ID length",
			args:       args{id: "short"},
			setup:      nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid ID format",
			args:       args{id: "invalid@id"},
			setup:      nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	urlService, _, err := setup()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}
	handler := NewURLHandlers(urlService)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.args.id = tt.setup(urlService)
			}
			req := httptest.NewRequest("GET", "/"+tt.args.id, nil)
			w := httptest.NewRecorder()
			router := chi.NewRouter()
			router.Get("/{id}", handler.Redirect)

			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if tt.wantStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.wantStatus, res.StatusCode)
				assert.NotEmpty(t, res.Header.Get("Location"), "Location header should not be empty")
			} else {
				assert.Equal(t, tt.wantStatus, res.StatusCode)
			}
		})
	}
}

func TestURLHandlers_ShortenJSON(t *testing.T) {
	type args struct {
		contentType string
		body        string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantBody   string
	}{
		{
			name: "valid JSON request",
			args: args{
				contentType: "application/json",
				body:        `{"url": "http://example.com"}`,
			},
			wantStatus: http.StatusCreated,
			wantBody:   `http://localhost:8080/`,
		},
		{
			name: "invalid Content-Type",
			args: args{
				contentType: "text/plain",
				body:        `{"url": "http://example.com"}`,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid content type\n",
		},
		{
			name: "invalid JSON format",
			args: args{
				contentType: "application/json",
				body:        `{"invalid_json"`,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "cannot decode request\n",
		},
	}

	urlService, _, err := setup()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	handler := NewURLHandlers(urlService)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(tt.args.body))
			req.Header.Set("Content-Type", tt.args.contentType)
			w := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Post("/api/shorten", handler.ShortenJSON)
			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			body, _ := io.ReadAll(res.Body)
			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantStatus == http.StatusCreated {
				var resp map[string]string
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err, "Response should be valid JSON")

				resultURL, ok := resp["result"]
				assert.True(t, ok, `"result" key should exist in the response`)
				assert.True(t, strings.HasPrefix(resultURL, tt.wantBody),
					"Expected result URL to start with %s, but got %s", tt.wantBody, resultURL)
			} else {
				assert.Equal(t, tt.wantBody, string(body))
			}
		})
	}
}
