package custom

import (
	"net/http"
	"time"

	"github.com/GlebRadaev/shlink/internal/logger"
)

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func RequestMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logger.NewLogger()
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

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
