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
		{
			name:            "Compressed request with invalid gzip data", // Новый тест
			handler:         http.HandlerFunc(okHandler),
			acceptEncoding:  "gzip",
			contentEncoding: "gzip",
			requestBody:     "Invalid gzip data",            // Некорректные данные
			expectedBody:    "",                             // Ожидаем пустое тело
			expectedStatus:  http.StatusInternalServerError, // Ошибка 500
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.contentEncoding == "gzip" {
				if tt.name == "Compressed request with invalid gzip data" {
					// Для некорректных данных просто передаем их как есть
					reqBody = strings.NewReader(tt.requestBody)
				} else {
					reqBody = gzipCompress(tt.requestBody)
				}
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
				if err != nil {
					// Если ошибка при декомпрессии, то тело должно быть пустым
					body = []byte{}
					assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
				} else {
					defer gzr.Close()
					body, err = io.ReadAll(gzr)
					assert.NoError(t, err, "ошибка при чтении сжатого тела ответа")
				}
			} else {
				body, err = io.ReadAll(res.Body)
				assert.NoError(t, err, "ошибка при чтении тела ответа")
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, string(body))
		})
	}
}

func TestCompressWriter_Header(t *testing.T) {
	rec := httptest.NewRecorder()
	compressWriter := &compressWriter{w: rec}

	rec.Header().Set("Content-Type", "application/json")
	rec.Header().Set("X-Custom-Header", "value")

	headers := compressWriter.Header()
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	assert.Equal(t, "value", headers.Get("X-Custom-Header"))

}

func TestCompressReader_Read(t *testing.T) {
	originalData := "Hello, world!"
	compressedData := gzipCompress(originalData)
	zr, err := gzip.NewReader(bytes.NewReader(compressedData.Bytes()))
	assert.NoError(t, err)
	reader := &compressReader{zr: zr}

	buf := make([]byte, len(originalData))

	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		t.Errorf("unexpected error: %v", err)
	}

	assert.Equal(t, len(originalData), n)
	assert.Equal(t, originalData, string(buf))
}

func BenchmarkCompressMiddleware(b *testing.B) {
	handler := http.HandlerFunc(okHandler)
	data := strings.Repeat("Hello, Benchmark!", 1000)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	req.Header.Set("Accept-Encoding", "gzip")

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		middleware := CompressMiddleware(handler)
		middleware.ServeHTTP(rec, req)
		res := rec.Result()
		_, err := io.ReadAll(res.Body)
		if err != nil {
			b.Fatalf("failed to read response body: %v", err)
		}
		res.Body.Close()
	}
}

func BenchmarkCompressReader(b *testing.B) {
	data := strings.Repeat("Hello, Benchmark!", 1000)
	compressed := gzipCompress(data)

	for i := 0; i < b.N; i++ {
		zr, err := gzip.NewReader(bytes.NewReader(compressed.Bytes()))
		if err != nil {
			b.Fatalf("Ошибка инициализации gzip.Reader: %v", err)
		}
		buf := make([]byte, len(data))
		_, _ = io.ReadFull(zr, buf)
		zr.Close()
	}
}
