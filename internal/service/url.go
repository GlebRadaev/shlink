package service

import (
	"errors"
	"fmt"

	"github.com/GlebRadaev/shlink/internal/config"
	repository "github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/utils"
)

const (
	MaxIDLength  = 8
	MaxURLLength = 2048
)

// URLService handles the business logic for shortening URLs
type URLService struct {
	storage repository.Repository
	config  *config.Config
}

// NewURLService creates a new URLService
func NewURLService(storage repository.Repository, config *config.Config) *URLService {
	return &URLService{
		storage: storage,
		config:  config,
	}
}

// ShortenURL shortens a given URL and returns the short version
func (s *URLService) Shorten(url string) (string, error) {
	_, err := utils.ValidateURL(url)
	if err != nil {
		return "", err
	}
	shortID := utils.Generate(MaxIDLength)
	err = s.storage.AddURL(shortID, url)
	if err != nil {
		return "", err
	}
	shortID = fmt.Sprintf("%s/%s", s.config.BaseURL, shortID)
	return shortID, nil
}

// GetOriginal retrieves the original URL by the short ID
func (s *URLService) GetOriginal(id string) (string, error) {
	if !utils.IsValidID(id, MaxIDLength) {
		return "", errors.New("invalid ID")
	}

	url, found := s.storage.Get(id)
	if !found {
		return "", errors.New("URL not found")
	}

	return url, nil
}
