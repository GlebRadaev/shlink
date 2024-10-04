package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestURLHandlers_Shorten(t *testing.T) {
	type fields struct {
		urlService *service.URLService
		config     *config.Config
	}
	type args struct {
		contentType string
		body        string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		setup      func(service *service.URLService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "valid URL",
			fields: fields{
				urlService: service.NewURLService(repository.NewMemoryStorage()),
				config:     &config.Config{BaseURL: "http://localhost"},
			},
			args: args{
				contentType: "text/plain",
				body:        "http://example.com",
			},
			setup: func(svc *service.URLService) {
			},
			wantStatus: http.StatusCreated,
			wantBody:   "http://localhost/shortID",
		},
		{
			name: "invalid Content-Type",
			fields: fields{
				urlService: service.NewURLService(repository.NewMemoryStorage()),
				config:     &config.Config{BaseURL: "http://localhost"},
			},
			args: args{
				contentType: "application/json",
				body:        "http://example.com",
			},
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid content type\n",
		},
		{
			name: "URL too long",
			fields: fields{
				urlService: service.NewURLService(repository.NewMemoryStorage()),
				config:     &config.Config{BaseURL: "http://localhost"},
			},
			args: args{
				contentType: "text/plain",
				body:        strings.Repeat("a", 2048+1),
			},
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   "URL is too long\n",
		},
		{
			name: "invalid URL format",
			fields: fields{
				urlService: service.NewURLService(repository.NewMemoryStorage()),
				config:     &config.Config{BaseURL: "http://localhost"},
			},
			args: args{
				contentType: "text/plain",
				body:        "invalid-url",
			},
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid URL format\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(tt.fields.urlService)
			}

			handler := &URLHandlers{
				urlService: tt.fields.urlService,
				config:     tt.fields.config,
			}

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
				assert.Contains(t, string(body), tt.fields.config.BaseURL+"/")
			} else {
				assert.Equal(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestURLHandlers_Redirect(t *testing.T) {
	type fields struct {
		urlService *service.URLService
		config     *config.Config
	}
	type args struct {
		id string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		setup      func(service *service.URLService) string
		wantStatus int
		wantHeader string
	}{
		{
			name: "valid ID",
			fields: fields{
				urlService: service.NewURLService(repository.NewMemoryStorage()),
				config:     &config.Config{},
			},
			args: args{id: "validID"},
			setup: func(svc *service.URLService) string {
				id, _ := svc.Shorten("http://example.com")
				return id
			},
			wantStatus: http.StatusTemporaryRedirect,
			wantHeader: "http://example.com",
		},
		{
			name: "ID not found",
			fields: fields{
				urlService: service.NewURLService(repository.NewMemoryStorage()),
				config:     &config.Config{},
			},
			args:       args{id: "notFoundID"},
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantHeader: "",
		},
		{
			name: "invalid ID length",
			fields: fields{
				urlService: service.NewURLService(repository.NewMemoryStorage()),
				config:     &config.Config{},
			},
			args:       args{id: "short"},
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantHeader: "",
		},
		{
			name: "invalid ID format",
			fields: fields{
				urlService: service.NewURLService(repository.NewMemoryStorage()),
				config:     &config.Config{},
			},
			args:       args{id: "invalid@id"},
			setup:      nil,
			wantStatus: http.StatusBadRequest,
			wantHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.args.id = tt.setup(tt.fields.urlService)
			}

			handler := &URLHandlers{
				urlService: tt.fields.urlService,
				config:     tt.fields.config,
			}

			r := chi.NewRouter()
			r.Get("/{id}", handler.Redirect)

			req := httptest.NewRequest("GET", "/"+tt.args.id, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.wantHeader, res.Header.Get("Location"))
			}
		})
	}
}
