package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type URLShortener struct {
	shortUrls map[string]string
}

var (
	urlShortener = &URLShortener{shortUrls: make(map[string]string)}
	cfg          *config.Config
)

const (
	maxURLLength = 2048
	idLength     = 8
)
const (
	errorURLTooLong         = "url is too long"
	errorInvalidContentType = "invalid content type. Only text/plain is allowed"
	errorInvalidRequestBody = "invalid request body"
	errorInvalidURLFormat   = "invalid URL format"
	errorInvalidID          = "invalid ID"
	errorInvalidIDFormat    = "invalid ID format"
	errorInvalidIDLength    = "invalid ID length"
)

func generateID() string {
	const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, idLength)
	charLen := len(characters)
	for i := range id {
		id[i] = characters[rand.Intn(charLen)]
	}
	return string(id)
}

func isValidURL(url string) error {
	validPrefixes := []string{"http://", "https://"}
	isValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(url, prefix) {
			isValidPrefix = true
			break
		}
	}
	if len(url) > maxURLLength {
		return errors.New(errorURLTooLong)
	}
	if !isValidPrefix {
		return errors.New(errorInvalidURLFormat)
	}
	if strings.Contains(url, " ") || strings.Contains(url, "#") || strings.Contains(url, "%") {
		return errors.New(errorInvalidURLFormat)
	}
	urlWithoutPrefix := strings.TrimPrefix(url, "http://")
	urlWithoutPrefix = strings.TrimPrefix(urlWithoutPrefix, "https://")
	if len(urlWithoutPrefix) == 0 || !strings.Contains(urlWithoutPrefix, ".") {
		return errors.New(errorInvalidURLFormat)
	}
	return nil
}

func isValidID(id string) error {
	if len(id) != idLength {
		return errors.New(errorInvalidIDLength)
	}
	validCharacters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for _, char := range id {
		if !strings.ContainsRune(validCharacters, char) {
			return errors.New(errorInvalidIDFormat)
		}
	}
	return nil
}

func (us *URLShortener) ShortenURL(url string) (string, error) {
	id := generateID()
	us.shortUrls[id] = url
	return fmt.Sprintf("%s/%s", cfg.BaseURL, id), nil
}

func (us *URLShortener) GetURL(id string) (string, bool) {
	url, ok := us.shortUrls[id]
	return url, ok
}

func respondWithError(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusBadRequest)
}

func shortenURL(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	parts := strings.Split(contentType, ";")
	if parts[0] != "text/plain" {
		respondWithError(w, errorInvalidContentType)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		respondWithError(w, errorInvalidRequestBody)
		return
	}
	url := string(body)

	if err := isValidURL(url); err != nil {
		respondWithError(w, err.Error())
		return
	}
	shortenedURL, err := urlShortener.ShortenURL(url)
	if err != nil {
		respondWithError(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenedURL))
}

func redirectURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := isValidID(id); err != nil {
		respondWithError(w, err.Error())
		return
	}
	url, ok := urlShortener.GetURL(id)
	if !ok {
		respondWithError(w, errorInvalidID)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		if duration > 60*time.Second {
			log.Printf("Request is too slow: %s %s is completed in %v", r.Method, r.URL.Path, duration)
		}
	})
}

func main() {
	cfg := config.ParseFlags()

	r := chi.NewRouter()
	r.Use(RequestMiddleware)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post(`/`, shortenURL)
	r.Get(`/{id}`, redirectURL)

	log.Printf("Server is running on %s", cfg.ServerAddress)
	err := http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		log.Println("Failed to start server on", cfg.ServerAddress)
		log.Fatal(err)
	}
}
