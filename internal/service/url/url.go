package url

import (
	"errors"
	"fmt"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/service/backup"
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
	backup     *backup.BackupService
	memoryRepo interfaces.Repository
}

// NewURLService creates a new URLService
func NewURLService(config *config.Config, log *logger.Logger, backup *backup.BackupService, memoryRepo interfaces.Repository) *URLService {
	return &URLService{
		log:        log.Named("URLService"),
		config:     config,
		backup:     backup,
		memoryRepo: memoryRepo,
	}
}

func (s *URLService) LoadData() error {
	data, err := s.backup.LoadData()
	if err != nil {
		return err
	}
	for shortURL, originalURL := range data {
		_ = s.memoryRepo.AddURL(shortURL, originalURL)
	}
	s.log.Info("Data successfully loaded from backup.")
	return nil
}

func (s *URLService) SaveData() error {
	data := s.memoryRepo.GetAll()
	if err := s.backup.SaveData(data); err != nil {
		return err
	}
	s.log.Info("Data successfully saved to backup.")
	return nil
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
