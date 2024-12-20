package url

import (
	"context"
	"errors"
	"fmt"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/service/backup"
	"github.com/GlebRadaev/shlink/internal/utils"

	"go.uber.org/zap"
)

const (
	MaxIDLength = 8
)

// URLService handles the business logic for shortening URLs
type URLService struct {
	log     *zap.SugaredLogger
	config  *config.Config
	backup  backup.IBackupService
	urlRepo interfaces.IURLRepository
}

// NewURLService creates a new URLService
func NewURLService(config *config.Config, log *logger.Logger, backup backup.IBackupService, urlRepo interfaces.IURLRepository) *URLService {
	return &URLService{
		log:     log.Named("URLService"),
		config:  config,
		backup:  backup,
		urlRepo: urlRepo,
	}
}

func (s *URLService) LoadData(ctx context.Context) error {
	data, err := s.backup.LoadData()
	if err != nil {
		return err
	}
	for shortURL, originalURL := range data {
		modelURL := &model.URL{
			ShortID:     shortURL,
			OriginalURL: originalURL,
		}
		_, _ = s.urlRepo.Insert(ctx, modelURL)
	}
	return nil
}

func (s *URLService) SaveData(ctx context.Context) error {
	urls, err := s.urlRepo.List(ctx)
	if err != nil {
		return err
	}
	data := make(map[string]string)
	for _, url := range urls {
		data[url.ShortID] = url.OriginalURL
	}
	if err := s.backup.SaveData(data); err != nil {
		return err
	}
	return nil
}

// ShortenURL shortens a given URL and returns the short version
func (s *URLService) Shorten(ctx context.Context, url string) (string, error) {
	s.log.Infof("Attempting to shorten URL: %s", url)
	_, err := utils.ValidateURL(url)
	if err != nil {
		s.log.Warnf("Invalid URL: %s, error: %v", url, err)
		return "", err
	}
	generateID := utils.Generate(MaxIDLength)
	// Создаем объект URL модели
	modelURL := model.URL{
		ShortID:     generateID,
		OriginalURL: url,
	}
	newURL, err := s.urlRepo.Insert(ctx, &modelURL)
	if err != nil {
		s.log.Errorf("Failed to add URL to memory repository: %v", err)
		return "", err
	}
	if newURL.ShortID != generateID {
		s.log.Infof("URL already exists: %s -> %s", newURL.OriginalURL, newURL.ShortID)
		return fmt.Sprintf("%s/%s", s.config.BaseURL, newURL.ShortID), errors.New("conflict: URL already shortened")
	}
	s.log.Infof("Successfully shortened URL: %s -> %s", newURL.OriginalURL, newURL.ShortID)
	shortID := fmt.Sprintf("%s/%s", s.config.BaseURL, newURL.ShortID)
	return shortID, nil
}

// ShortenURL shortens a given URL and returns the short version
func (s *URLService) ShortenList(ctx context.Context, data dto.BatchShortenRequestDTO) (dto.BatchShortenResponseDTO, error) {
	resultData := make([]dto.BatchShortenResponse, 0, len(data))
	insertData := make([]*model.URL, 0, len(data))
	for _, dataInfo := range data {
		_, err := utils.ValidateURL(dataInfo.OriginalURL)
		if err != nil {
			s.log.Warnf("Invalid URL: %s, error: %v", dataInfo.OriginalURL, err)
			continue
		}
		modelURL := model.URL{
			ShortID:     utils.Generate(MaxIDLength),
			OriginalURL: dataInfo.OriginalURL,
		}
		insertData = append(insertData, &modelURL)
		resultData = append(resultData, dto.BatchShortenResponse{
			CorrelationID: dataInfo.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.config.BaseURL, modelURL.ShortID),
		})
	}
	if len(insertData) > 0 {
		_, err := s.urlRepo.InsertList(ctx, insertData)
		if err != nil {
			s.log.Errorf("Failed to add URL to memory repository: %v", err)
			return nil, err
		}
	}
	return resultData, nil
}

// GetOriginal retrieves the original URL by the short ID
func (s *URLService) GetOriginal(ctx context.Context, id string) (string, error) {
	s.log.Infof("Retrieving original URL for ID: %s", id)
	if !utils.IsValidID(id, MaxIDLength) {
		s.log.Warnf("Invalid ID: %s", id)
		return "", errors.New("invalid ID")
	}
	url, err := s.urlRepo.FindByID(ctx, id)
	if err != nil {
		s.log.Errorf("Error retrieving URL for ID %s: %v", id, err)
		return "", err
	}
	if url == nil {
		s.log.Errorf("URL not found for ID %s", id)
		return "", errors.New("URL not found")
	}
	s.log.Infof("Found URL for ID %s: %s", id, url.OriginalURL)
	return url.OriginalURL, nil
}
