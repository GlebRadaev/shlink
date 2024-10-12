package custom

import (
	"log"
	"net/http"
	"time"
)

func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		if duration > 60*time.Second {
			log.Printf("Request is too slow: %s %s completed in %v", r.Method, r.URL.Path, duration)
		}
	})
}
