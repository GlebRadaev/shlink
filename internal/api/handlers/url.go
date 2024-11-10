package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/service"
	"github.com/GlebRadaev/shlink/internal/utils"

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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)

		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	shortID, err := h.urlService.Shorten(r.Context(), string(body))
	if err != nil {
		fmt.Println(err)

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortID))
	if err != nil {
		fmt.Println(err)

		http.Error(w, "Failed to write response", http.StatusBadRequest)
		return
	}
}

// Redirect handles the request to redirect to the original URL
func (h *URLHandlers) Redirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	originalURL, err := h.urlService.GetOriginal(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URLHandlers) ShortenJSON(w http.ResponseWriter, r *http.Request) {
	if err := utils.ValidateContentType(w, r, "application/json"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var data dto.ShortenRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "cannot decode request", http.StatusBadRequest)
		return
	}
	if data.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}
	shortID, err := h.urlService.Shorten(r.Context(), data.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(dto.ShortenResponseDTO{Result: shortID}); err != nil {
		http.Error(w, "Error encoding response", http.StatusBadRequest)
		return
	}
}
