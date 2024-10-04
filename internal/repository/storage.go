package repository

import "errors"

type MemoryStorage struct {
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (s *MemoryStorage) Save(shortID string, longURL string) error {
	if shortID == "" {
		return errors.New("shortID cannot be empty")
	}
	s.data[shortID] = longURL
	return nil
}

func (s *MemoryStorage) Find(shortID string) (string, bool) {
	url, exists := s.data[shortID]
	return url, exists
}
