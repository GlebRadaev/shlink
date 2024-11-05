package inmemory

import (
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

func (s *MemoryStorage) AddURL(key, value string) error {
	if key == "" {
		return errors.New("shortID cannot be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *MemoryStorage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, exists := s.data[key]
	return url, exists
}

func (s *MemoryStorage) GetAll() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}
