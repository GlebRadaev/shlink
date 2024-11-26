package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service"
	"github.com/GlebRadaev/shlink/internal/service/url"
	"github.com/GlebRadaev/shlink/internal/taskmanager"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type mockErrorReader struct{}

func (m *mockErrorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("mock read error")
}

func setupURL(ctx context.Context) (*url.URLService, *config.Config, error) {
	if cfgTest == nil {
		var err error
		cfgTest, err = config.ParseAndLoadConfig()
		if err != nil {
			return nil, nil, err
		}
	}
	log, _ := logger.NewLogger("info")
	pool := taskmanager.NewWorkerPool(ctx, 10, 1)
	repositories := repository.NewRepositoryFactory(ctx, cfgTest, log)
	services := service.NewServiceFactory(ctx, cfgTest, log, pool, repositories)
	return services.URLService, cfgTest, nil
}

func TestURLHandlers_Shorten(t *testing.T) {
	ctx := context.Background()
	type args struct {
		contentType string
		body        string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantBody   string
		mockReader io.Reader
	}{
		{
			name: "valid URL",
			args: args{
				contentType: "text/plain",
				body:        fmt.Sprintf("http://example.com?test=%d", time.Now().UnixNano()),
			},
			wantStatus: http.StatusCreated,
			wantBody:   "http://localhost/shortID",
			mockReader: strings.NewReader(fmt.Sprintf("http://example.com?test=%d", time.Now().UnixNano())),
		},
		{
			name: "invalid URL format (URL too long)",
			args: args{
				contentType: "text/plain",
				body:        strings.Repeat("a", 2048+1),
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid URL format\n",
			mockReader: strings.NewReader(strings.Repeat("a", 2048+1)),
		},
		{
			name: "invalid URL format",
			args: args{
				contentType: "text/plain",
				body:        "invalid-url",
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid URL format\n",
			mockReader: strings.NewReader("invalid-url"),
		},
		{
			name: "failed to read request body",
			args: args{
				contentType: "text/plain",
				body:        "http://example.com",
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Failed to read request body\n",
			mockReader: &mockErrorReader{},
		},
	}

	urlService, cfg, err := setupURL(ctx)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	handler := NewURLHandlers(urlService)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", tt.mockReader)
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
	ctx := context.Background()

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
				url, _ := service.Shorten(ctx, "userID", "http://example.com")
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

	urlService, _, err := setupURL(ctx)
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
	ctx := context.Background()
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
				body:        fmt.Sprintf(`{"url": "http://example.com?test=%d"}`, time.Now().UnixNano()),
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
		{
			name: "empty request body",
			args: args{
				contentType: "application/json",
				body:        "",
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "cannot decode request\n",
		},
		{
			name: "missing URL",
			args: args{
				contentType: "application/json",
				body:        `{"url": ""}`,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "url is required\n",
		},
	}

	urlService, _, err := setupURL(ctx)
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

func TestURLHandlers_ShortenJSONBatch(t *testing.T) {
	ctx := context.Background()
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
			name: "valid JSON batch request",
			args: args{
				contentType: "application/json",
				body:        `[{"correlation_id": "1", "original_url": "http://example.com"}, {"correlation_id": "2", "original_url": "http://another-example.com"}]`,
			},
			wantStatus: http.StatusCreated,
			wantBody:   `[{"correlation_id":"1","shortened_url":"http://localhost/shortID1"},{"correlation_id":"2","shortened_url":"http://localhost/shortID2"}]`,
		},
		{
			name: "invalid JSON format",
			args: args{
				contentType: "application/json",
				body:        `[{"correlation_id": "1", "original_url": "http://example.com"}, {"correlation_id": "2", "original_url": "http://another-example.com"}`, // неполный JSON
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "cannot decode request\n",
		},
		{
			name: "invalid Content-Type",
			args: args{
				contentType: "text/plain",
				body:        `[{"correlation_id": "1", "original_url": "http://example.com"}]`,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid content type\n",
		},
		{
			name: "empty request body",
			args: args{
				contentType: "application/json",
				body:        "",
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "cannot decode request\n",
		},
		{
			name: "missing URL in batch",
			args: args{
				contentType: "application/json",
				body:        `[{"correlation_id": "1", "original_url": ""}]`,
			},
			wantStatus: http.StatusCreated,
			wantBody:   "",
		},
	}

	urlService, _, err := setupURL(ctx)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}
	handler := NewURLHandlers(urlService)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/shorten/batch", strings.NewReader(tt.args.body))
			req.Header.Set("Content-Type", tt.args.contentType)
			w := httptest.NewRecorder()

			handler.ShortenJSONBatch(w, req)

			res := w.Result()
			defer res.Body.Close()

			body, _ := io.ReadAll(res.Body)

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			matched, err := regexp.MatchString(tt.wantBody, string(body))
			assert.NoError(t, err)
			assert.True(t, matched, fmt.Sprintf("Expected body to match regex: %s, but got: %s", tt.wantBody, string(body)))
		})
	}
}
