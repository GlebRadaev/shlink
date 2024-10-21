package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func largeBodyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(bytes.Repeat([]byte("A"), 1024*1024))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func TestRequestMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.Handler
		method         string
		expectedStatus int
		expectedSize   int
	}{
		{
			name:           "GET request with OK status",
			handler:        http.HandlerFunc(okHandler),
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedSize:   len("OK"),
		},
		{
			name:           "GET request with 404 status",
			handler:        http.HandlerFunc(notFoundHandler),
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
			expectedSize:   0,
		},
		{
			name:           "POST request not allowed",
			handler:        http.HandlerFunc(methodNotAllowedHandler),
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedSize:   0,
		},
		{
			name:           "Large body response",
			handler:        http.HandlerFunc(largeBodyHandler),
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedSize:   1024 * 1024,
		},
		{
			name:           "Empty response body",
			handler:        http.HandlerFunc(notFoundHandler),
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
			expectedSize:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			rec := httptest.NewRecorder()

			middleware := RequestMiddleware(tt.handler)
			middleware.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode, "unexpected status code")
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err, "failed to read response body")
			assert.Equal(t, tt.expectedSize, len(body), "unexpected response size")
		})
	}
}
