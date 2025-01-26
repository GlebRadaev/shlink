// Package inmemory provides an in-memory implementation of the IURLRepository
// interface for managing URL data storage. It supports basic CRUD operations
// on URLs with synchronization support using read/write locks.
//
// The MemoryStorage struct holds the data in a map and ensures thread-safe
// access to it. It supports operations like Insert, InsertList, FindByID, and
// FindListByUserID, and it can delete a list of URLs based on user ID and short IDs.
//
// Example usage:
//
//	storage := inmemory.NewMemoryStorage()
//	url, err := storage.Insert(context.Background(), &model.URL{OriginalURL: "https://example.com"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Use other methods to interact with URLs in memory.
package inmemory

import (
	"context"
	"sync"

	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/utils"
)

// MemoryStorage is an in-memory storage implementation of IURLRepository.
// It uses a map for storage and a mutex for thread-safe access.
type MemoryStorage struct {
	data map[string]model.URL // Map of shortID to URL
	mu   sync.RWMutex         // Read/Write mutex for synchronization
}

// NewMemoryStorage creates a new instance of MemoryStorage that implements
// the IURLRepository interface. It initializes the internal map for URL storage.
func NewMemoryStorage() interfaces.IURLRepository {
	return &MemoryStorage{
		data: make(map[string]model.URL),
	}
}

// Insert stores a URL in memory or returns the existing one if the original
// URL is already in the storage. It generates a new ShortID if needed.
func (s *MemoryStorage) Insert(ctx context.Context, url *model.URL) (*model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Check if the URL already exists
	for _, storedURL := range s.data {
		if storedURL.OriginalURL == url.OriginalURL {
			url.ShortID = storedURL.ShortID
			return url, nil
		}
	}
	// Ensure unique ShortID
	if _, exists := s.data[url.ShortID]; exists {
		url.ShortID = utils.Generate(8)
	}
	s.data[url.ShortID] = *url
	return url, nil
}

// InsertList stores a list of URLs in memory, reusing ShortID for duplicate
// original URLs and generating new ShortIDs if needed.
func (s *MemoryStorage) InsertList(ctx context.Context, urls []*model.URL) ([]*model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, url := range urls {
		// Check for existing URLs and reuse their ShortID
		for _, storedURL := range s.data {
			if storedURL.OriginalURL == url.OriginalURL {
				url.ShortID = storedURL.ShortID
				break
			}
		}
		// Ensure unique ShortID
		if _, exists := s.data[url.ShortID]; exists {
			url.ShortID = utils.Generate(8)
		}
		s.data[url.ShortID] = *url
	}
	return urls, nil
}

// FindByID retrieves a URL from the storage by its ShortID. Returns nil if
// the ShortID does not exist.
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

// FindListByUserID retrieves a list of URLs associated with a specific user
// by their UserID.
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

// List retrieves a list of all URLs stored in memory. The returned list will
// contain shallow copies of the URLs to avoid external modifications.
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

// Ping checks if the storage is accessible. This can be used to verify
// the health of the storage.
func (s *MemoryStorage) Ping(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

// DeleteListByUserIDAndShortIDs deletes a list of URLs from memory by
// UserID and their associated ShortIDs. It updates the URLs' state in memory.
func (s *MemoryStorage) DeleteListByUserIDAndShortIDs(ctx context.Context, userID string, shortIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, shortID := range shortIDs {
		if url, exists := s.data[shortID]; exists && url.UserID == userID {
			s.data[shortID] = url
		}
	}
	return nil
}
