package inmemory_test

import (
	"context"
	"testing"

	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/repository/inmemory"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_Insert(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()

	tests := []struct {
		name    string
		shortID string
		longURL string
		wantErr error
	}{
		{
			name:    "valid shortID and longURL",
			shortID: "abc123",
			longURL: "http://example.com",
			wantErr: nil,
		},
		{
			name:    "empty shortID",
			shortID: "",
			longURL: "http://example.com",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &model.URL{
				ShortID:     tt.shortID,
				OriginalURL: tt.longURL,
			}
			savedURL, err := storage.Insert(ctx, url)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				foundURL, err := storage.FindByID(ctx, tt.shortID)
				assert.NoError(t, err)
				assert.Equal(t, savedURL.OriginalURL, foundURL.OriginalURL, "Expected the longURL to match the saved value")
			}
		})
	}
}

func TestMemoryStorage_FindById(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()

	tests := []struct {
		name       string
		shortID    string
		storedData map[string]string
		wantURL    string
		wantExists bool
	}{
		{
			name:       "shortID exists",
			shortID:    "abc123",
			storedData: map[string]string{"abc123": "http://example.com"},
			wantURL:    "http://example.com",
			wantExists: true,
		},
		{
			name:       "shortID does not exist",
			shortID:    "xyz789",
			storedData: map[string]string{"abc123": "http://example.com"},
			wantURL:    "",
			wantExists: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for shortID, originalURL := range tt.storedData {
				_, err := storage.Insert(ctx, &model.URL{ShortID: shortID, OriginalURL: originalURL})
				if err != nil {
					t.Errorf("Insert() returned an error: %v", err)
					continue
				}
			}

			foundURL, err := storage.FindByID(ctx, tt.shortID)
			assert.NoError(t, err)
			if tt.wantExists {
				assert.Equal(t, tt.wantURL, foundURL.OriginalURL, "Expected the URL to match the stored value")
			} else {
				assert.Nil(t, foundURL, "Expected no URL to be found")
			}
		})
	}
}

func TestMemoryStorage_List(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()

	if _, err := storage.Insert(ctx, &model.URL{ShortID: "abc123", OriginalURL: "http://example.com"}); err != nil {
		t.Errorf("Insert() returned an error: %v", err)
	}
	if _, err := storage.Insert(ctx, &model.URL{ShortID: "xyz789", OriginalURL: "http://another-example.com"}); err != nil {
		t.Errorf("Insert() returned an error: %v", err)
	}

	t.Run("get all URLs", func(t *testing.T) {
		expectedURLs := []*model.URL{
			{ShortID: "abc123", OriginalURL: "http://example.com"},
			{ShortID: "xyz789", OriginalURL: "http://another-example.com"},
		}
		allURLs, err := storage.List(ctx)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expectedURLs, allURLs, "List should return the correct list of URLs")
	})
}
