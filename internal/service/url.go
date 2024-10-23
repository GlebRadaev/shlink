package service

import (
	"errors"
	"fmt"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/utils"
)

const (
	MaxIDLength = 8
)

// URLService handles the business logic for shortening URLs
type URLService struct {
	memoryRepo interfaces.Repository
	fileRepo   interfaces.Repository
	config     *config.Config
}

// NewURLService creates a new URLService
func NewURLService(memoryRepo, fileRepo interfaces.Repository, config *config.Config) *URLService {
	copyData(fileRepo, memoryRepo)
	return &URLService{
		memoryRepo: memoryRepo,
		fileRepo:   fileRepo,
		config:     config,
	}
}

func copyData(from, to interfaces.Repository) {
	for shortURL, url := range from.GetAll() {
		if err := to.AddURL(shortURL, url); err != nil {
			fmt.Println(err)
		}
	}
}

// ShortenURL shortens a given URL and returns the short version
func (s *URLService) Shorten(url string) (string, error) {
	_, err := utils.ValidateURL(url)
	if err != nil {
		return "", err
	}
	shortID := utils.Generate(MaxIDLength)
	err = s.memoryRepo.AddURL(shortID, url)
	if err != nil {
		return "", err
	}
	err = s.fileRepo.AddURL(shortID, url)
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

	url, found := s.memoryRepo.Get(id)
	if !found {
		return "", errors.New("URL not found")
	}

	return url, nil
}
