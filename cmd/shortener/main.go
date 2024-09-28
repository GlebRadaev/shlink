package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

var (
    baseURL = "http://localhost"
    port    = "8080"
    shortUrls = make(map[string]string)
)

const (
    urlTooLongMessage     = "URL is too long"
    invalidContentTypeMsg  = "Invalid content type. Only text/plain is allowed"
    invalidRequestBodyMsg  = "Invalid request body"
    invalidURLFormatMsg    = "Invalid URL format"
    invalidIDMsg           = "Invalid ID"
    invalidIDLengthMsg     = "Invalid ID length"
)

func generateID() string {
    const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
    id := make([]byte, 8)
    charLen := len(characters)
    for i := range id {
        id[i] = characters[rand.Intn(charLen)]
    }
    return string(id)
}

func isValidURL(url string) bool {
    validPrefixes := []string{"http://", "https://"}
    isValidPrefix := false
    for _, prefix := range validPrefixes {
        if strings.HasPrefix(url, prefix) {
            isValidPrefix = true
            break
        }
    }
    if !isValidPrefix {
        return false
    }
    if strings.Contains(url, " ") || strings.Contains(url, "#") || strings.Contains(url, "%") {
        return false
    }
    urlWithoutPrefix := strings.TrimPrefix(url, "http://")
    urlWithoutPrefix = strings.TrimPrefix(urlWithoutPrefix, "https://")
    if len(urlWithoutPrefix) == 0 || !strings.Contains(urlWithoutPrefix, ".") {
        return false
    }
    return true
}

func respondWithError(w http.ResponseWriter, message string) {
    http.Error(w, message, http.StatusBadRequest)
}

func shortenURL(w http.ResponseWriter, r *http.Request) {
    contentType := r.Header.Get("Content-Type")
    parts := strings.Split(contentType, ";")
    if parts[0] != "text/plain" {
        respondWithError(w, invalidContentTypeMsg)
        return
    }
    body, err := io.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
        respondWithError(w, invalidRequestBodyMsg)
        return
    }
    url := string(body)

    if !isValidURL(url) {
        respondWithError(w, invalidURLFormatMsg)
        return
    }
    if len(url) > 2048 {
        respondWithError(w, urlTooLongMessage)
        return
    }
    if !isValidURL(url) {
        respondWithError(w, invalidURLFormatMsg)
        return
    }
    id := generateID()
    shortUrls[id] = url
    shortenedURL := fmt.Sprintf("%s:%s/%s", baseURL, port, id)
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte(shortenedURL))
}

func redirectURL(w http.ResponseWriter, r *http.Request) {
    id := strings.TrimPrefix(r.URL.Path, "/")
    if len(id) != 8 {
        respondWithError(w, invalidIDLengthMsg)
        return
    }
    url, ok := shortUrls[id]
    if !ok {
        respondWithError(w, invalidIDMsg)
        return
    }
    w.Header().Set("Location", url)
    w.WriteHeader(http.StatusTemporaryRedirect)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
        shortenURL(w, r)
    case http.MethodGet:
        redirectURL(w, r)
    default:
        respondWithError(w, "Method not allowed")
    }
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc(`/`, handleRequest)

    log.Printf("Server is running on %s:%s", baseURL, port)
    err := http.ListenAndServe(":"+port, mux)
    if err != nil {
        log.Fatal(err)
    }
}