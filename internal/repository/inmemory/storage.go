package inmemory

import (
	"errors"

	"github.com/GlebRadaev/shlink/internal/interfaces"
)

type MemoryStorage struct {
	data map[string]string
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
	s.data[key] = value
	return nil
}

func (s *MemoryStorage) Get(key string) (string, bool) {
	url, exists := s.data[key]

	return url, exists
}

func (s *MemoryStorage) GetAll() map[string]string {
	return s.data
}
