package handlers

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/GlebRadaev/shlink/internal/service"
	"github.com/go-chi/chi/v5"
)

// URLHandlers defines the handlers for URL shortening
type URLHandlers struct {
	urlService *service.URLService
}

// NewURLHandlers creates a new instance of URLHandlers
func NewURLHandlers(urlService *service.URLService) *URLHandlers {
	return &URLHandlers{urlService: urlService}
}

// Shorten handles the request to shorten a URL
func (h *URLHandlers) Shorten(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !isValidContentType(contentType) {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	url := string(body)
	shortURL, err := h.urlService.Shorten(url)
	log.Print(err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// Redirect handles the request to redirect to the original URL
func (h *URLHandlers) Redirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	contentType := r.Header.Get("Content-Type")
	if !isValidContentType(contentType) {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	originalURL, err := h.urlService.GetOriginal(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Helper function to validate content type.
func isValidContentType(contentType string) bool {
	parts := strings.Split(contentType, ";")
	return parts[0] == "text/plain"
}
