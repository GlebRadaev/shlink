package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/GlebRadaev/shlink/internal/dto"
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

func (h *URLHandlers) ShortenJSON(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	parts := strings.Split(contentType, ";")
	if len(parts) == 0 || !strings.Contains(parts[0], "application/json") {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}
	if r.Body == nil {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	var req dto.ShortenJSONRequestDTO
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "Cannot decode request", http.StatusBadRequest)
		return
	}
	shortID, err := h.urlService.Shorten(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := dto.ShortenJSONResponseDTO{
		Result: shortID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		http.Error(w, "Error encoding response", http.StatusBadRequest)
		return
	}
}
