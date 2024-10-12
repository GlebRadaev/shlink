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
	parts := strings.Split(contentType, ";")
	if len(parts) == 0 || !strings.Contains(parts[0], "text/plain") {
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
	shortID, err := h.urlService.Shorten(url)
	log.Print(err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortID))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusBadRequest)
		return
	}
}

// Redirect handles the request to redirect to the original URL
func (h *URLHandlers) Redirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	originalURL, err := h.urlService.GetOriginal(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
