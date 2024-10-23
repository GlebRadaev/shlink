package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
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

func gzipCompress(data string) *bytes.Buffer {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(data))
	if err != nil {
		log.Fatalf("Ошибка сжатия: %v", err)
	}
	gz.Close()
	return &buf
}

func TestCompressMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		handler         http.Handler
		acceptEncoding  string
		contentEncoding string
		requestBody     string
		expectedBody    string
		expectedStatus  int
	}{
		{
			name:            "Uncompressed request and response",
			handler:         http.HandlerFunc(okHandler),
			acceptEncoding:  "",
			contentEncoding: "",
			requestBody:     "Hello, world!",
			expectedBody:    "OK",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "Compressed request, uncompressed response",
			handler:         http.HandlerFunc(okHandler),
			acceptEncoding:  "",
			contentEncoding: "gzip",
			requestBody:     "Hello, compressed!",
			expectedBody:    "OK",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "Uncompressed request, compressed response",
			handler:         http.HandlerFunc(okHandler),
			acceptEncoding:  "gzip",
			contentEncoding: "",
			requestBody:     "",
			expectedBody:    "OK",
			expectedStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.contentEncoding == "gzip" {
				reqBody = gzipCompress(tt.requestBody)
			} else {
				reqBody = strings.NewReader(tt.requestBody)
			}
			req := httptest.NewRequest(http.MethodPost, "/", reqBody)
			req.Header.Set("Content-Encoding", tt.contentEncoding)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)

			rec := httptest.NewRecorder()
			middleware := CompressMiddleware(tt.handler)
			middleware.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			var body []byte
			var err error
			if tt.acceptEncoding == "gzip" {
				gzr, err := gzip.NewReader(res.Body)
				assert.NoError(t, err, "ошибка при создании gzip.Reader")
				defer gzr.Close()
				body, err = io.ReadAll(gzr)
				assert.NoError(t, err, "ошибка при чтении сжатого тела ответа")
			} else {
				body, err = io.ReadAll(res.Body)
				assert.NoError(t, err, "ошибка при чтении тела ответа")
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, string(body))
		})
	}
}
