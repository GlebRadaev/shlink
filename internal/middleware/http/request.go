package middleware

import (
	"net/http"
	"time"

	"github.com/GlebRadaev/shlink/internal/logger"
)

type (
	// responseData holds information about the response status and size.
	responseData struct {
		status int // HTTP response status code
		size   int // Size of the response body
	}

	// loggingResponseWriter wraps the original ResponseWriter to capture
	// response status and size for logging purposes.
	loggingResponseWriter struct {
		http.ResponseWriter
		*responseData
	}
)

// RequestMiddleware logs the details of incoming requests and their responses.
func RequestMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, _ := logger.NewLogger("info")
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		logger.Named("Request").Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
		)
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		logger.Named("Response").Infoln(
			"status", responseData.status,
			"size", responseData.size,
			"duration", duration,
		)
	})
}

// Write intercepts the write operation to capture the response body size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader captures the response status code.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
