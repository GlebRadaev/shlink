package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GlebRadaev/shlink/internal/middleware"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	r := chi.NewMux()
	middleware.Middleware(r)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		realIP := r.Header.Get("X-Forwarded-For")
		w.Header().Set("X-Real-IP", realIP)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode, "unexpected status code")
	body := rec.Body.String()
	assert.Equal(t, "OK", body, "unexpected response body")
	assert.Equal(t, "192.0.2.1", res.Header.Get("X-Real-IP"), "unexpected real IP")

	t.Run("recoverer middleware", func(t *testing.T) {
		r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})
		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code, "unexpected status code on panic")
	})
}
