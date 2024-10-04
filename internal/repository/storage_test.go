package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_Save(t *testing.T) {
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
			wantErr: errors.New("shortID cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemoryStorage()

			err := storage.Save(tt.shortID, tt.longURL)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				savedURL, exists := storage.Find(tt.shortID)
				assert.True(t, exists, "Expected the shortID to be saved, but it was not found")
				assert.Equal(t, tt.longURL, savedURL, "Expected the longURL to match the saved value")
			}
		})
	}
}

func TestMemoryStorage_Find(t *testing.T) {
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
			storage := &MemoryStorage{
				data: tt.storedData,
			}
			url, exists := storage.Find(tt.shortID)
			assert.Equal(t, tt.wantExists, exists)
			if tt.wantExists {
				assert.Equal(t, tt.wantURL, url, "Expected the URL to match the stored value")
			} else {
				assert.Empty(t, url, "Expected no URL to be found")
			}
		})
	}
}
