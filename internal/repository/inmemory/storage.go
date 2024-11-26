package inmemory

import (
	"context"
	"sync"

	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/utils"
)

type MemoryStorage struct {
	data map[string]model.URL
	mu   sync.RWMutex
}

func NewMemoryStorage() interfaces.IURLRepository {
	return &MemoryStorage{
		data: make(map[string]model.URL),
	}
}

func (s *MemoryStorage) Insert(ctx context.Context, url *model.URL) (*model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, storedURL := range s.data {
		if storedURL.OriginalURL == url.OriginalURL {
			url.ShortID = storedURL.ShortID
			return url, nil
		}
	}
	if _, exists := s.data[url.ShortID]; exists {
		url.ShortID = utils.Generate(8)
	}
	s.data[url.ShortID] = *url
	return url, nil
}

func (s *MemoryStorage) InsertList(ctx context.Context, urls []*model.URL) ([]*model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, url := range urls {
		for _, storedURL := range s.data {
			if storedURL.OriginalURL == url.OriginalURL {
				url.ShortID = storedURL.ShortID
				break
			}
		}
		if _, exists := s.data[url.ShortID]; exists {
			url.ShortID = utils.Generate(8)
		}
		s.data[url.ShortID] = *url
	}
	return urls, nil
}

func (s *MemoryStorage) FindByID(ctx context.Context, shortID string) (*model.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	url, exists := s.data[shortID]
	if !exists {
		return nil, nil
	}
	return &url, nil
}

func (s *MemoryStorage) FindListByUserID(ctx context.Context, userID string) ([]*model.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var result []*model.URL
	for _, storedURL := range s.data {
		if storedURL.UserID == userID {
			urlCopy := storedURL
			result = append(result, &urlCopy)
		}
	}
	return result, nil
}

func (s *MemoryStorage) List(ctx context.Context) ([]*model.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var result []*model.URL
	for _, storedURL := range s.data {
		urlCopy := model.URL{
			ID:          storedURL.ID,
			ShortID:     storedURL.ShortID,
			OriginalURL: storedURL.OriginalURL,
			UserID:      storedURL.UserID,
			CreatedAt:   storedURL.CreatedAt,
		}
		result = append(result, &urlCopy)
	}
	return result, nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (s *MemoryStorage) DeleteListByUserIDAndShortIDs(ctx context.Context, userID string, shortIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, shortID := range shortIDs {
		if url, exists := s.data[shortID]; exists && url.UserID == userID {
			// url.IsDeleted = true
			s.data[shortID] = url
		}
	}
	return nil
}
