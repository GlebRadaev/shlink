package filestorage

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileStorage_AddURL(t *testing.T) {
	fileRepo := NewFileStorage("test_data.txt")
	defer os.Remove("test_data.txt")

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
			err := fileRepo.AddURL(tt.shortID, tt.longURL)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				savedURL, exists := fileRepo.Get(tt.shortID)
				assert.True(t, exists, "Expected shortID to be saved, but it was not found")
				assert.Equal(t, tt.longURL, savedURL)
			}
		})
	}
}

func TestFileStorage_Get(t *testing.T) {
	fileRepo := NewFileStorage("test_data.txt")
	defer os.Remove("test_data.txt")

	tests := []struct {
		name       string
		shortID    string
		storedData map[string]string
		wantURL    string
		wantExists bool
	}{
		{
			name:    "shortID exists",
			shortID: "abc123",
			storedData: map[string]string{
				"abc123": "http://example.com",
			},
			wantURL:    "http://example.com",
			wantExists: true,
		},
		{
			name:    "shortID does not exist",
			shortID: "xyz789",
			storedData: map[string]string{
				"abc123": "http://example.com",
			},
			wantURL:    "",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, url := range tt.storedData {
				_ = fileRepo.AddURL(key, url)
			}
			url, exists := fileRepo.Get(tt.shortID)
			assert.Equal(t, tt.wantExists, exists)
			assert.Equal(t, tt.wantURL, url)
		})
	}
}

func TestFileStorage_GetAll(t *testing.T) {
	fileRepo := NewFileStorage("test_data.txt")
	defer os.Remove("test_data.txt")

	_ = fileRepo.AddURL("abc123", "http://example.com")
	_ = fileRepo.AddURL("xyz789", "http://another-example.com")

	expected := map[string]string{
		"abc123": "http://example.com",
		"xyz789": "http://another-example.com",
	}
	t.Run("get all URLs", func(t *testing.T) {
		allURLs := fileRepo.GetAll()
		assert.Equal(t, expected, allURLs)
	})
}
