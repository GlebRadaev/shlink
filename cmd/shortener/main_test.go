package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func assertResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedCode int, expectedError string) {
	assert.Equal(t, expectedCode, rr.Code)
	if expectedError != "" {
		assert.Equal(t, expectedError, strings.TrimSpace(rr.Body.String()))
	}
}

func setupRouter() *chi.Mux {
	cfg = &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post(`/`, shortenURL)
		r.Get(`/{id}`, redirectURL)
	})

	return r
}

func TestIsValidURL(t *testing.T) {
	testCases := []struct {
		name       string
		url        string
		shouldFail bool
	}{
		{"Valid HTTP URL", "http://example.com", false},
		{"Valid HTTPS URL", "https://example.com", false},
		{"Invalid FTP URL", "ftp://example.com", true},
		{"URL with space", "http://example .com", true},
		{"URL with hash symbol", "http://example.com#", true},
		{"Empty URL after protocol", "https://", true},
		{"URL without domain", "http://example", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := isValidURL(tc.url)
			if tc.shouldFail {
				assert.NotNil(t, err, "Expected error but got none")
			} else {
				assert.Nil(t, err, "Expected no error but got one")
			}
		})
	}
}

func TestShortenURL(t *testing.T) {
	tests := []struct {
		name          string
		contentType   string
		body          string
		expectedCode  int
		expectedError string
	}{
		{
			name:          "Invalid Content-Type",
			contentType:   "application/json",
			body:          "http://example.com",
			expectedCode:  http.StatusBadRequest,
			expectedError: errorInvalidContentType,
		},
		{
			name:          "Invalid URL Format",
			contentType:   "text/plain",
			body:          "http://example/com",
			expectedCode:  http.StatusBadRequest,
			expectedError: errorInvalidURLFormat,
		},
		{
			name:          "URL Too Long",
			contentType:   "text/plain",
			body:          "http://." + strings.Repeat("a", 2049),
			expectedCode:  http.StatusBadRequest,
			expectedError: errorURLTooLong,
		},
		{
			name:         "Successful URL Shortening",
			contentType:  "text/plain",
			body:         "http://example.com",
			expectedCode: http.StatusCreated,
		},
	}
	r := setupRouter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()

			r.Post("/", shortenURL)

			r.ServeHTTP(rr, req)
			assert.Equal(t, tt.expectedCode, rr.Code)
			assertResponse(t, rr, tt.expectedCode, tt.expectedError)
			if tt.expectedError == "" {
				body := rr.Body.String()
				expectedPrefix := fmt.Sprintf("%s/", cfg.BaseURL)
				assert.Contains(t, body, expectedPrefix)

				parts := strings.Split(strings.TrimSpace(body), "/")
				generatedID := parts[len(parts)-1]
				assert.Len(t, generatedID, 8, "Generated ID should be 8 characters long")
			}
		})
	}
}

func TestRedirectURL(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		expectedCode  int
		expectedError string
		expectedURL   string
	}{
		{
			name:         "Valid ID",
			expectedCode: http.StatusTemporaryRedirect,
			expectedURL:  "http://example.com",
		},
		{
			name:          "Invalid ID",
			id:            "12345678",
			expectedCode:  http.StatusBadRequest,
			expectedError: errorInvalidID,
		},
		{
			name:          "Invalid ID Length",
			id:            "short",
			expectedCode:  http.StatusBadRequest,
			expectedError: errorInvalidIDLength,
		},
	}
	r := setupRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Valid ID" {
				shortenReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com"))
				shortenReq.Header.Set("Content-Type", "text/plain")
				shortenResp := httptest.NewRecorder()
				r.Post("/", shortenURL)
				r.ServeHTTP(shortenResp, shortenReq)
				assert.Equal(t, http.StatusCreated, shortenResp.Code)
				shortenedURL := shortenResp.Body.String()
				tt.id = strings.TrimPrefix(shortenedURL, fmt.Sprintf("%s/", cfg.BaseURL))
			}

			req := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/{id}", redirectURL)

			r.ServeHTTP(rr, req)
			assert.Equal(t, tt.expectedCode, rr.Code)
			assertResponse(t, rr, tt.expectedCode, tt.expectedError)
			if tt.expectedError == "" {
				assert.Equal(t, tt.expectedURL, rr.Header().Get("Location"))
			}
		})
	}
}
