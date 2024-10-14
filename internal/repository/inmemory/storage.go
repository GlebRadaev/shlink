package inmemory

import (
	"errors"

	"github.com/GlebRadaev/shlink/internal/repository"
)

type MemoryStorage struct {
	data map[string]string
}

func NewMemoryStorage() repository.Repository {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (s *MemoryStorage) AddUrl(key, value string) error {
	if key == "" {
		return errors.New("shortID cannot be empty")
	}
	s.data[key] = value
	return nil
}

func (s *MemoryStorage) Get(key string) (string, bool) {
	url, exists := s.data[key]

	return url, exists
}
