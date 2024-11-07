package inmemory

import (
	"context"
	"errors"
	"sync"

	"github.com/GlebRadaev/shlink/internal/interfaces"
)

type MemoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage() interfaces.Repository {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (s *MemoryStorage) AddURL(ctx context.Context, key, value string) error {
	if key == "" {
		return errors.New("shortID cannot be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	s.data[key] = value
	return nil
}

func (s *MemoryStorage) Get(ctx context.Context, key string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	url, exists := s.data[key]
	return url, exists, nil
}

func (s *MemoryStorage) GetAll(ctx context.Context) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.data, nil
}
