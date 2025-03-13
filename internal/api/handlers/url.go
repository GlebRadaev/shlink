// Package handlers provides HTTP handlers for URL shortening operations,
// including creating short links, redirecting to original URLs, and managing user URLs.
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/service"
	"github.com/GlebRadaev/shlink/internal/utils"
)

// URLHandlers defines the handlers for URL shortening.
type URLHandlers struct {
	// urlService is the service that manages URL shortening and retrieval operations.
	urlService *service.URLService
}

// NewURLHandlers creates a new instance of URLHandlers.
func NewURLHandlers(urlService *service.URLService) *URLHandlers {
	return &URLHandlers{urlService: urlService}
}

// Shorten handles the request to shorten a URL.
func (h *URLHandlers) Shorten(w http.ResponseWriter, r *http.Request) {
	userID, _ := utils.GetOrSetUserIDFromCookie(w, r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	shortID, err := h.urlService.Shorten(r.Context(), userID, string(body))
	if err != nil {
		if strings.Contains(err.Error(), "conflict") {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(shortID))
			return
		}
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

// Redirect handles the request to redirect to the original URL.
func (h *URLHandlers) Redirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	originalURL, err := h.urlService.GetOriginal(r.Context(), id)
	log.Printf("Redirecting to: %s", originalURL)
	if err != nil {
		if err.Error() == "URL is deleted" {
			http.Error(w, err.Error(), http.StatusGone)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// ShortenJSON handles the request to shorten a single URL in JSON format.
func (h *URLHandlers) ShortenJSON(w http.ResponseWriter, r *http.Request) {
	userID, _ := utils.GetOrSetUserIDFromCookie(w, r)
	if err := utils.ValidateContentType(w, r, "application/json"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var data dto.ShortenJSONRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "cannot decode request", http.StatusBadRequest)
		return
	}
	if data.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	shortID, err := h.urlService.Shorten(r.Context(), userID, data.URL)
	if err != nil {
		if strings.Contains(err.Error(), "conflict") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(dto.ShortenJSONResponseDTO{Result: shortID})
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(dto.ShortenJSONResponseDTO{Result: shortID}); err != nil {
		http.Error(w, "Error encoding response", http.StatusBadRequest)
		return
	}
}

// ShortenJSONBatch handles the request to shorten multiple URLs provided in a batch JSON format.
func (h *URLHandlers) ShortenJSONBatch(w http.ResponseWriter, r *http.Request) {
	userID, _ := utils.GetOrSetUserIDFromCookie(w, r)
	if err := utils.ValidateContentType(w, r, "application/json"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var data dto.BatchShortenRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "cannot decode request", http.StatusBadRequest)
		return
	}

	shortenResults, err := h.urlService.ShortenList(r.Context(), userID, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(shortenResults); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetUserURLs retrieves the list of URLs associated with the authenticated user.
func (h *URLHandlers) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetOrSetUserIDFromCookie(w, r)
	if err != nil {
		http.Error(w, "Failed to get or set user ID", http.StatusInternalServerError)
		return
	}

	urls, err := h.urlService.GetUserURLs(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(urls) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(urls); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// DeleteUserURLs deletes a list of URLs associated with the authenticated user.
func (h *URLHandlers) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserIDFromCookie(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var data dto.DeleteURLRequestDTO
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.urlService.DeleteUserURLs(r.Context(), userID, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// GetStats retrieves the statistics for the URL shortening service.
func (h *URLHandlers) GetStats(w http.ResponseWriter, r *http.Request) {
	clientIP := r.Header.Get("X-Real-IP")
	if clientIP == "" {
		http.Error(w, "X-Real-IP header is missing", http.StatusForbidden)
		return
	}

	if !h.urlService.IsAllowed(clientIP) {
		http.Error(w, "Access forbidden", http.StatusForbidden)
		return
	}

	stats, err := h.urlService.GetStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
