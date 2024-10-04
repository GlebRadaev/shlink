package service

import (
	"errors"
	"strings"

	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/utils"
)

const (
	MaxIdLength  = 8
	MaxURLLength = 2048
)

// URLService handles the business logic for shortening URLs
type URLService struct {
	storage *repository.MemoryStorage
}

// NewURLService creates a new URLService
func NewURLService(storage *repository.MemoryStorage) *URLService {
	return &URLService{storage: storage}
}

// ShortenURL shortens a given URL and returns the short version
func (s *URLService) Shorten(url string) (string, error) {
	if len(url) > MaxURLLength {
		return "", errors.New("URL is too long")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", errors.New("invalid URL format")
	}
	urlWithoutPrefix := strings.TrimPrefix(url, "http://")
	urlWithoutPrefix = strings.TrimPrefix(urlWithoutPrefix, "https://")
	if strings.Contains(url, " ") || strings.Contains(url, "#") || strings.Contains(url, "%") || len(urlWithoutPrefix) == 0 || !strings.Contains(urlWithoutPrefix, ".") {
		return "", errors.New("invalid URL format")
	}
	id := utils.Generate(MaxIdLength)
	err := s.storage.Save(id, url)
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetOriginal retrieves the original URL by the short ID
func (s *URLService) GetOriginal(id string) (string, error) {
	if len(id) != MaxIdLength {
		return "", errors.New("invalid ID length")
	}
	validCharacters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for _, char := range id {
		if !strings.ContainsRune(validCharacters, char) {
			return "", errors.New("invalid ID format")
		}
	}
	url, found := s.storage.Find(id)
	if !found {
		return "", errors.New("URL not found")
	}

	return url, nil
}
