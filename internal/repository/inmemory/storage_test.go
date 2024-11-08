package inmemory

import (
	"context"
	"testing"

	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_Insert(t *testing.T) {
	ctx := context.Background()
	memoryRepo := NewMemoryStorage()
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
			savedURL, err := memoryRepo.Insert(ctx, url)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				foundURL, err := memoryRepo.FindByID(ctx, tt.shortID)
				assert.NoError(t, err)
				assert.Equal(t, savedURL.OriginalURL, foundURL.OriginalURL, "Expected the longURL to match the saved value")
			}
		})
	}
}

func TestMemoryStorage_FindById(t *testing.T) {
	ctx := context.Background()
	memoryRepo := NewMemoryStorage()
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
			for shortID, longURL := range tt.storedData {
				_, _ = memoryRepo.Insert(ctx, &model.URL{ShortID: shortID, OriginalURL: longURL})
			}
			foundURL, err := memoryRepo.FindByID(ctx, tt.shortID)
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
	ctx := context.Background()
	memoryRepo := NewMemoryStorage()
	_, _ = memoryRepo.Insert(ctx, &model.URL{ShortID: "abc123", OriginalURL: "http://example.com"})
	_, _ = memoryRepo.Insert(ctx, &model.URL{ShortID: "xyz789", OriginalURL: "http://another-example.com"})
	t.Run("get all URLs", func(t *testing.T) {
		allURLs, err := memoryRepo.List(ctx)
		assert.NoError(t, err)

		expected := []*model.URL{
			{ShortID: "abc123", OriginalURL: "http://example.com"},
			{ShortID: "xyz789", OriginalURL: "http://another-example.com"},
		}

		assert.ElementsMatch(t, expected, allURLs, "List should return the correct list of URLs")
	})
}
