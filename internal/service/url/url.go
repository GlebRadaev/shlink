// Package url provides services for URL shortening and management.
package url

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/service/backup"
	"github.com/GlebRadaev/shlink/internal/taskmanager"
	"github.com/GlebRadaev/shlink/internal/utils"

	"go.uber.org/zap"
)

// MaxIDLength defines the maximum length for a shortened URL ID.
const (
	MaxIDLength = 8
)

// URLService handles the business logic for shortening URLs
// and interacts with repositories, backups, and tasks related to URL management.
type URLService struct {
	log      *zap.SugaredLogger        // Logger for the service
	config   *config.Config            // Configuration settings for the service
	taskPool *taskmanager.WorkerPool   // Worker pool for handling tasks
	backup   backup.IBackupService     // Backup service for saving and loading URL data
	urlRepo  interfaces.IURLRepository // Repository for interacting with stored URLs
}

// NewURLService creates a new instance of URLService with the specified configurations
// and registers the task handler for deleting URLs.
func NewURLService(
	config *config.Config,
	log *logger.Logger,
	pool *taskmanager.WorkerPool,
	backup backup.IBackupService,
	urlRepo interfaces.IURLRepository,
) *URLService {
	service := &URLService{
		log:      log.Named("URLService"),
		config:   config,
		backup:   backup,
		urlRepo:  urlRepo,
		taskPool: pool,
	}
	pool.RegisterHandler("delete_urls_task", service.ProcessDeleteURLsTask)
	return service
}

// LoadData loads previously backed-up URL data and inserts them into the repository.
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

// SaveData retrieves all URLs and backs them up to persistent storage.
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

// ProcessDeleteURLsTask processes a task that deletes a list of URLs for a specific user.
func (s *URLService) ProcessDeleteURLsTask(ctx context.Context, task taskmanager.Task) error {
	deleteTask, ok := task.(taskmanager.DeleteTask)
	if !ok {
		return fmt.Errorf("invalid task type: expected DeleteTask")
	}
	s.log.Infof("Starting delete task for userID=%s with %d URLs", deleteTask.UserID, len(deleteTask.URLs))

	const batchSize = 10
	var wg sync.WaitGroup
	errChan := make(chan error, len(deleteTask.URLs)/batchSize+1)
	successChan := make(chan string, len(deleteTask.URLs)/batchSize+1)
	for i := 0; i < len(deleteTask.URLs); i += batchSize {
		end := i + batchSize
		if end > len(deleteTask.URLs) {
			end = len(deleteTask.URLs)
		}
		batch := deleteTask.URLs[i:end]
		s.log.Infof("Processing batch for userID=%s: %v", deleteTask.UserID, batch)

		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()
			err := s.urlRepo.DeleteListByUserIDAndShortIDs(ctx, deleteTask.UserID, batch)
			if err != nil {
				errChan <- fmt.Errorf("error deleting batch for userID=%s: %v", deleteTask.UserID, err)
			} else {
				successChan <- fmt.Sprintf("batch deleted for userID=%s: %v", deleteTask.UserID, batch)
			}
		}(batch)
	}

	go func() {
		wg.Wait()
		close(errChan)
		close(successChan)
	}()

	for {
		select {
		case err, ok := <-errChan:
			if ok && err != nil {
				s.log.Errorf("Error in delete task for userID=%s: %v", deleteTask.UserID, err)
				return err
			}
		case success, ok := <-successChan:
			if ok {
				s.log.Infof("Success: %s", success)
			}
		}
		if len(errChan) == 0 && len(successChan) == 0 {
			break
		}
	}
	s.log.Infof("Completed delete task for userID=%s", deleteTask.UserID)
	return nil
}

// Shorten shortens a given URL and returns the corresponding short version.
func (s *URLService) Shorten(ctx context.Context, userID string, url string) (string, error) {
	s.log.Infof("Attempting to shorten URL: %s", url)
	_, err := utils.ValidateURL(url)
	if err != nil {
		s.log.Warnf("Invalid URL: %s, error: %v", url, err)
		return "", err
	}
	generateID := utils.Generate(MaxIDLength)
	modelURL := model.URL{
		ShortID:     generateID,
		OriginalURL: url,
		UserID:      userID,
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

// ShortenList shortens a batch of URLs and returns the corresponding short versions.
func (s *URLService) ShortenList(ctx context.Context, userID string, data dto.BatchShortenRequestDTO) (dto.BatchShortenResponseDTO, error) {
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
			UserID:      userID,
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

// GetOriginal retrieves the original URL associated with the given short ID.
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
	if url.DeletedFlag {
		s.log.Errorf("URL is deleted for ID %s", id)
		return "", errors.New("URL is deleted")
	}
	s.log.Infof("Found URL for ID %s: %s", id, url.OriginalURL)
	return url.OriginalURL, nil
}

// GetUserURLs retrieves all URLs shortened by a user.
func (s *URLService) GetUserURLs(ctx context.Context, userID string) (dto.GetUserURLsResponseDTO, error) {
	urls, err := s.urlRepo.FindListByUserID(ctx, userID)
	if err != nil {
		s.log.Errorf("Error getting URLs for user ID %s: %v", userID, err)
		return nil, err
	}
	if len(urls) == 0 {
		s.log.Errorf("URL not found for user ID %s", userID)
		return dto.GetUserURLsResponseDTO{}, nil
	}
	var responseDTO dto.GetUserURLsResponseDTO
	for _, url := range urls {
		responseDTO = append(responseDTO, dto.GetUserURLsResponse{
			ShortURL:    fmt.Sprintf("%s/%s", s.config.BaseURL, url.ShortID),
			OriginalURL: url.OriginalURL,
		})
	}
	return responseDTO, nil
}

// DeleteUserURLs schedules a task to delete multiple URLs for a specific user.
func (s *URLService) DeleteUserURLs(ctx context.Context, userID string, urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	task := taskmanager.DeleteTask{
		UserID: userID,
		URLs:   urls,
	}
	err := s.taskPool.Enqueue(ctx, task)
	if err != nil {
		s.log.Errorf("Failed to enqueue task: %v", err)
	}

	s.log.Infof("Starting delete task for userID=%s with %d URLs", task.UserID, len(task.URLs))
	return nil
}

// const batchSize = 10
// var wg sync.WaitGroup
// errChan := make(chan error, len(urls)/batchSize+1)
// successChan := make(chan string, len(urls)/batchSize+1)
// for i := 0; i < len(urls); i += batchSize {
// 	end := i + batchSize
// 	if end > len(urls) {
// 		end = len(urls)
// 	}
// 	batch := urls[i:end]
// 	s.log.Infof("Processing batch for userID=%s: %v", userID, batch)

// 	wg.Add(1)
// 	go func(batch []string) {
// 		defer wg.Done()
// 		err := s.urlRepo.DeleteListByUserIDAndShortIDs(ctx, userID, batch)
// 		if err != nil {
// 			errChan <- fmt.Errorf("Error deleting batch for userID=%s: %v", userID, err)
// 		} else {
// 			successChan <- fmt.Sprintf("Batch deleted for userID=%s: %v", userID, batch)
// 		}
// 	}(batch)
// }

// go func() {
// 	wg.Wait()
// 	close(errChan)
// 	close(successChan)
// }()

// for {
// 	select {
// 	case err, ok := <-errChan:
// 		if ok && err != nil {
// 			s.log.Errorf("Error in delete task for userID=%s: %v", userID, err)
// 			return err
// 		}
// 	case success, ok := <-successChan:
// 		if ok {
// 			s.log.Infof("Success: %s", success)
// 		}
// 	}
// 	if len(errChan) == 0 && len(successChan) == 0 {
// 		break
// 	}
// }
// s.log.Infof("Completed delete task for userID=%s", userID)

// return nil
// }
