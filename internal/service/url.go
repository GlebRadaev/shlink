package service

import (
	"errors"
	"fmt"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/utils"
	"go.uber.org/zap"
)

const (
	MaxIDLength = 8
)

// URLService handles the business logic for shortening URLs
type URLService struct {
	log        *zap.SugaredLogger
	config     *config.Config
	memoryRepo interfaces.Repository
	fileRepo   interfaces.Repository
}

// NewURLService creates a new URLService
func NewURLService(config *config.Config, memoryRepo, fileRepo interfaces.Repository) *URLService {
	log := logger.NewLogger()
	copyData(fileRepo, memoryRepo)
	return &URLService{
		log:        log.Named("URLService"),
		config:     config,
		memoryRepo: memoryRepo,
		fileRepo:   fileRepo,
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
	s.log.Infof("Attempting to shorten URL: %s", url)
	_, err := utils.ValidateURL(url)
	if err != nil {
		s.log.Warnf("Invalid URL: %s, error: %v", url, err)
		return "", err
	}
	shortID := utils.Generate(MaxIDLength)
	err = s.memoryRepo.AddURL(shortID, url)
	if err != nil {
		s.log.Errorf("Failed to add URL to memory repository: %v", err)
		return "", err
	}
	err = s.fileRepo.AddURL(shortID, url)
	if err != nil {
		s.log.Errorf("Failed to add URL to file repository: %v", err)
		return "", err
	}
	s.log.Infof("Successfully shortened URL: %s -> %s", url, shortID)
	shortID = fmt.Sprintf("%s/%s", s.config.BaseURL, shortID)
	return shortID, nil
}

// GetOriginal retrieves the original URL by the short ID
func (s *URLService) GetOriginal(id string) (string, error) {
	s.log.Infof("Retrieving original URL for ID: %s", id)
	if !utils.IsValidID(id, MaxIDLength) {
		s.log.Warnf("Invalid ID: %s", id)
		return "", errors.New("invalid ID")
	}

	url, found := s.memoryRepo.Get(id)
	if !found {
		s.log.Warnf("URL not found for ID: %s", id)
		return "", errors.New("URL not found")
	}
	s.log.Infof("Found URL for ID %s: %s", id, url)
	return url, nil
}
