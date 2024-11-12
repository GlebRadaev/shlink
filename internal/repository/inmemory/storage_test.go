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
		{
			name:    "context cancelled",
			shortID: "abc123",
			longURL: "http://example.com",
			wantErr: context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &model.URL{
				ShortID:     tt.shortID,
				OriginalURL: tt.longURL,
			}
			if tt.wantErr == context.Canceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			savedURL, err := storage.Insert(ctx, url)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				if tt.shortID == "" {
					assert.NotEmpty(t, savedURL.ShortID, "Expected a generated ShortID for an empty input")
				} else {
					assert.Equal(t, tt.shortID, savedURL.ShortID)
				}
				foundURL, err := storage.FindByID(ctx, savedURL.ShortID)
				assert.NoError(t, err)
				assert.Equal(t, savedURL.OriginalURL, foundURL.OriginalURL, "Expected the longURL to match the saved value")
			}
		})
	}
}

func TestMemoryStorage_InsertList(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()

	tests := []struct {
		name    string
		urls    []*model.URL
		wantErr error
	}{
		{
			name: "valid list of URLs",
			urls: []*model.URL{
				{ShortID: "abc123", OriginalURL: "http://example.com"},
				{ShortID: "xyz789", OriginalURL: "http://another-example.com"},
			},
			wantErr: nil,
		},
		{
			name:    "empty list of URLs",
			urls:    []*model.URL{},
			wantErr: nil,
		},
		{
			name:    "context cancelled",
			urls:    []*model.URL{{ShortID: "abc123", OriginalURL: "http://example.com"}},
			wantErr: context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr == context.Canceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			savedURLs, err := storage.InsertList(ctx, tt.urls)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				for _, savedURL := range savedURLs {
					foundURL, err := storage.FindByID(ctx, savedURL.ShortID)
					assert.NoError(t, err)
					assert.Equal(t, savedURL.OriginalURL, foundURL.OriginalURL, "Expected the longURL to match the saved value")
				}
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
		wantErr    error
	}{
		{
			name:       "shortID exists",
			shortID:    "abc123",
			storedData: map[string]string{"abc123": "http://example.com"},
			wantURL:    "http://example.com",
			wantExists: true,
			wantErr:    nil,
		},
		{
			name:       "shortID does not exist",
			shortID:    "xyz789",
			storedData: map[string]string{"abc123": "http://example.com"},
			wantURL:    "",
			wantExists: false,
			wantErr:    nil,
		},
		{
			name:       "context cancelled",
			shortID:    "abc123",
			storedData: map[string]string{"abc123": "http://example.com"},
			wantURL:    "",
			wantExists: false,
			wantErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr == context.Canceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			for shortID, originalURL := range tt.storedData {
				_, err := storage.Insert(ctx, &model.URL{ShortID: shortID, OriginalURL: originalURL})
				if err != nil && err != tt.wantErr {
					t.Errorf("Insert() returned an error: %v", err)
				}
			}
			foundURL, err := storage.FindByID(ctx, tt.shortID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				if tt.wantExists {
					assert.Equal(t, tt.wantURL, foundURL.OriginalURL, "Expected the URL to match the stored value")
				} else {
					assert.Nil(t, foundURL, "Expected no URL to be found")
				}
			}
		})
	}
}

func TestMemoryStorage_List(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()

	tests := []struct {
		name          string
		shortID       string
		originalURL   string
		expectedURLs  []*model.URL
		expectedError error
	}{
		{
			name:        "get all URLs",
			shortID:     "abc123",
			originalURL: "http://example.com",
			expectedURLs: []*model.URL{
				{ShortID: "abc123", OriginalURL: "http://example.com"},
				{ShortID: "xyz789", OriginalURL: "http://another-example.com"},
			},
			expectedError: nil,
		},
		{
			name:          "context cancelled",
			shortID:       "abc123",
			originalURL:   "http://example.com",
			expectedURLs:  nil,
			expectedError: context.Canceled,
		},
	}

	if _, err := storage.Insert(ctx, &model.URL{ShortID: "abc123", OriginalURL: "http://example.com"}); err != nil {
		t.Errorf("Insert() returned an error: %v", err)
	}
	if _, err := storage.Insert(ctx, &model.URL{ShortID: "xyz789", OriginalURL: "http://another-example.com"}); err != nil {
		t.Errorf("Insert() returned an error: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "context cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			allURLs, err := storage.List(ctx)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.ElementsMatch(t, tt.expectedURLs, allURLs, "List should return the correct list of URLs")
		})
	}
}

func TestMemoryStorage_Ping(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()

	tests := []struct {
		name          string
		ctx           context.Context
		expectedError error
	}{
		{
			name:          "ping successful",
			ctx:           ctx,
			expectedError: nil,
		},
		{
			name:          "context cancelled",
			ctx:           func() context.Context { ctx, cancel := context.WithCancel(ctx); cancel(); return ctx }(), // создаем контекст с отменой
			expectedError: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.Ping(tt.ctx)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "Expected error to match")
			} else {
				assert.NoError(t, err, "Expected no error")
			}
		})
	}
}
