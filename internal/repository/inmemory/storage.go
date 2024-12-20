package inmemory

import (
	"context"
	"sync"

	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/utils"
)

type MemoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage() interfaces.IURLRepository {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (s *MemoryStorage) Insert(ctx context.Context, url *model.URL) (*model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for shortID, originalURL := range s.data {
		if originalURL == url.OriginalURL {
			url.ShortID = shortID
			return url, nil
		}
	}
	if _, exists := s.data[url.ShortID]; exists {
		url.ShortID = utils.Generate(8)
	}
	s.data[url.ShortID] = url.OriginalURL
	return url, nil
}

func (s *MemoryStorage) InsertList(ctx context.Context, urls []*model.URL) ([]*model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, url := range urls {
		s.data[url.ShortID] = url.OriginalURL
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
	return &model.URL{OriginalURL: url}, nil
}

func (s *MemoryStorage) List(ctx context.Context) ([]*model.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	urls := make([]*model.URL, 0, len(s.data))
	for shortID, originalURL := range s.data {
		urls = append(urls, &model.URL{
			ShortID:     shortID,
			OriginalURL: originalURL,
		})
	}
	return urls, nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
