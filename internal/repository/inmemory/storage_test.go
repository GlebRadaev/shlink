package inmemory_test

import (
	"context"
	"testing"
	"time"

	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/repository/inmemory"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_CountURLs(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()

	tests := []struct {
		name          string
		storedData    map[string]model.URL
		expectedCount int
		expectedError error
	}{
		{
			name:          "Empty storage",
			storedData:    map[string]model.URL{},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name: "Successful CountURLs",
			storedData: map[string]model.URL{
				"abc123": {ShortID: "abc123", OriginalURL: "http://example.com", UserID: "user1"},
				"xyz789": {ShortID: "xyz789", OriginalURL: "http://another.com", UserID: "user2"},
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "Context cancelled",
			storedData: map[string]model.URL{
				"abc123": {ShortID: "abc123", OriginalURL: "http://example.com", UserID: "user1"},
			},
			expectedCount: 0,
			expectedError: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, url := range tt.storedData {
				_, err := storage.Insert(ctx, &url)
				assert.NoError(t, err)
			}

			if tt.expectedError == context.Canceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			count, err := storage.CountURLs(ctx)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}
		})
	}
}

func TestMemoryStorage_CountUsers(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()

	tests := []struct {
		name          string
		storedData    map[string]model.URL
		expectedCount int
		expectedError error
	}{
		{
			name:          "Empty storage",
			storedData:    map[string]model.URL{},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name: "Duplicate users",
			storedData: map[string]model.URL{
				"abc123": {ShortID: "abc123", OriginalURL: "http://example.com", UserID: "user1"},
				"xyz789": {ShortID: "xyz789", OriginalURL: "http://another.com", UserID: "user1"},
			},
			expectedCount: 1,
			expectedError: nil,
		},
		{
			name: "Successful CountUsers",
			storedData: map[string]model.URL{
				"abc123": {ShortID: "abc123", OriginalURL: "http://example1.com", UserID: "user0"},
				"xyz789": {ShortID: "xyz789", OriginalURL: "http://another.com", UserID: "user2"},
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "Context cancelled",
			storedData: map[string]model.URL{
				"abc123": {ShortID: "abc123", OriginalURL: "http://example.com", UserID: "user1"},
			},
			expectedCount: 0,
			expectedError: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, url := range tt.storedData {
				_, err := storage.Insert(ctx, &url)
				assert.NoError(t, err)
			}

			if tt.expectedError == context.Canceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			count, err := storage.CountUsers(ctx)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}
		})
	}
}

func TestMemoryStorage_Insert(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()

	tests := []struct {
		name              string
		shortID           string
		originalURL       string
		userID            string
		expectedError     error
		expectedShort     string
		expectedUserID    string
		expectedCreatedAt time.Time
	}{
		{
			name:              "valid shortID and unique OriginalURL",
			shortID:           "abc123",
			originalURL:       "http://example.com",
			userID:            "user1",
			expectedError:     nil,
			expectedShort:     "abc123",
			expectedUserID:    "user1",
			expectedCreatedAt: time.Now(),
		},
		{
			name:              "duplicate OriginalURL",
			shortID:           "xyz789",
			originalURL:       "http://example.com",
			userID:            "user2",
			expectedError:     nil,
			expectedShort:     "abc123",
			expectedUserID:    "user2",
			expectedCreatedAt: time.Now(),
		},
		{
			name:              "context cancelled",
			shortID:           "abc123",
			originalURL:       "http://example.com",
			userID:            "user3",
			expectedError:     context.Canceled,
			expectedShort:     "",
			expectedUserID:    "",
			expectedCreatedAt: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedError == context.Canceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			url := &model.URL{
				ShortID:     tt.shortID,
				OriginalURL: tt.originalURL,
				UserID:      tt.userID,
				CreatedAt:   time.Now(),
			}
			savedURL, err := storage.Insert(ctx, url)
			assert.Equal(t, tt.expectedError, err)

			if tt.expectedError == nil {
				assert.Equal(t, tt.expectedShort, savedURL.ShortID)
				assert.Equal(t, tt.expectedUserID, savedURL.UserID)
				assert.WithinDuration(t, savedURL.CreatedAt, time.Now(), time.Second)
				foundURL, err := storage.FindByID(ctx, savedURL.ShortID)
				assert.NoError(t, err)
				assert.Equal(t, tt.originalURL, foundURL.OriginalURL)
				if tt.shortID == "" {
					assert.NotEmpty(t, savedURL.ShortID, "Expected generated ShortID for empty input")
				}
			}
		})
	}
}

func TestMemoryStorage_InsertList(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()

	tests := []struct {
		name              string
		urls              []*model.URL
		expectedError     error
		expectedUserIDs   []string
		expectedCreatedAt time.Time
	}{
		{
			name: "valid shortIDs and unique OriginalURLs",
			urls: []*model.URL{
				{
					ShortID:     "abc123",
					OriginalURL: "http://example.com",
					UserID:      "user1",
					CreatedAt:   time.Now(),
				},
				{
					ShortID:     "xyz789",
					OriginalURL: "http://another.com",
					UserID:      "user2",
					CreatedAt:   time.Now(),
				},
			},
			expectedError:     nil,
			expectedUserIDs:   []string{"user1", "user2"},
			expectedCreatedAt: time.Now(),
		},
		{
			name: "duplicate OriginalURLs",
			urls: []*model.URL{
				{
					ShortID:     "abc123",
					OriginalURL: "http://example.com",
					UserID:      "user1",
					CreatedAt:   time.Now(),
				},
				{
					ShortID:     "xyz789",
					OriginalURL: "http://example.com",
					UserID:      "user2",
					CreatedAt:   time.Now(),
				},
			},
			expectedError:     nil,
			expectedUserIDs:   []string{"user1", "user2"},
			expectedCreatedAt: time.Now(),
		},
		{
			name: "context cancelled",
			urls: []*model.URL{
				{
					ShortID:     "abc123",
					OriginalURL: "http://example.com",
					UserID:      "user3",
					CreatedAt:   time.Now(),
				},
			},
			expectedError:     context.Canceled,
			expectedUserIDs:   []string{},
			expectedCreatedAt: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedError == context.Canceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			savedURLs, err := storage.InsertList(ctx, tt.urls)
			assert.Equal(t, tt.expectedError, err)

			if tt.expectedError == nil {
				assert.Len(t, savedURLs, len(tt.urls))
				for i, savedURL := range savedURLs {
					assert.NotEmpty(t, savedURL.ShortID)
					assert.Equal(t, tt.expectedUserIDs[i], savedURL.UserID)
					assert.WithinDuration(t, savedURL.CreatedAt, time.Now(), time.Second)
					foundURL, err := storage.FindByID(ctx, savedURL.ShortID)
					assert.NoError(t, err)
					assert.Equal(t, tt.urls[i].OriginalURL, foundURL.OriginalURL)
				}
			}
		})
	}
}

func TestMemoryStorage_FindById(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name       string
		shortID    string
		storedData map[string]model.URL
		wantURL    *model.URL
		wantExists bool
		wantErr    error
	}{
		{
			name:    "shortID exists",
			shortID: "abc123",
			storedData: map[string]model.URL{
				"abc123": {
					ID:          1,
					ShortID:     "abc123",
					OriginalURL: "http://example.com",
					UserID:      "user1",
					CreatedAt:   now,
				},
			},
			wantURL: &model.URL{
				ID:          1,
				ShortID:     "abc123",
				OriginalURL: "http://example.com",
				UserID:      "user1",
				CreatedAt:   now,
			},
			wantExists: true,
			wantErr:    nil,
		},
		{
			name:    "shortID does not exist",
			shortID: "xyz789",
			storedData: map[string]model.URL{
				"abc123": {
					ID:          1,
					ShortID:     "abc123",
					OriginalURL: "http://example.com",
					UserID:      "user1",
					CreatedAt:   now,
				},
			},
			wantURL:    nil,
			wantExists: false,
			wantErr:    nil,
		},
		{
			name:    "context cancelled",
			shortID: "abc123",
			storedData: map[string]model.URL{
				"abc123": {
					ID:          1,
					ShortID:     "abc123",
					OriginalURL: "http://example.com",
					UserID:      "user1",
					CreatedAt:   now,
				},
			},
			wantURL:    nil,
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
			for _, url := range tt.storedData {
				_, err := storage.Insert(ctx, &url)
				if err != nil && err != tt.wantErr {
					t.Fatalf("Insert() returned an error: %v", err)
				}
			}
			foundURL, err := storage.FindByID(ctx, tt.shortID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				if tt.wantExists {
					assert.NotNil(t, foundURL, "Expected a URL to be found")
					assert.Equal(t, tt.wantURL, foundURL, "Expected the URL to match the stored value")
				} else {
					assert.Nil(t, foundURL, "Expected no URL to be found")
				}
			}
		})
	}
}
func TestMemoryStorage_FindListByUserID(t *testing.T) {
	storage := inmemory.NewMemoryStorage()
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name       string
		userID     string
		storedData map[string]model.URL
		wantURLs   []*model.URL
		wantErr    error
	}{
		{
			name:   "User ID does not exist",
			userID: "user456",
			storedData: map[string]model.URL{
				"abc123": {ID: 1, ShortID: "abc123", OriginalURL: "http://example.com", UserID: "user123", CreatedAt: now},
			},
			wantURLs: []*model.URL{},
			wantErr:  nil,
		},
		{
			name:   "User ID exists",
			userID: "user123",
			storedData: map[string]model.URL{
				"abc123": {ID: 1, ShortID: "abc123", OriginalURL: "http://example.com", UserID: "user123", CreatedAt: now},
				"xyz789": {ID: 2, ShortID: "xyz789", OriginalURL: "http://another.com", UserID: "user123", CreatedAt: now},
			},
			wantURLs: []*model.URL{
				{ID: 1, ShortID: "abc123", OriginalURL: "http://example.com", UserID: "user123", CreatedAt: now},
				{ID: 2, ShortID: "xyz789", OriginalURL: "http://another.com", UserID: "user123", CreatedAt: now},
			},
			wantErr: nil,
		},
		{
			name:   "Context cancelled",
			userID: "user123",
			storedData: map[string]model.URL{
				"abc123": {ID: 1, ShortID: "abc123", OriginalURL: "http://example.com", UserID: "user123", CreatedAt: now},
			},
			wantURLs: []*model.URL{},
			wantErr:  context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr == context.Canceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			for _, url := range tt.storedData {
				_, err := storage.Insert(ctx, &url)
				if err != nil && err != tt.wantErr {
					t.Fatalf("Insert() returned an error: %v", err)
				}
			}

			foundURLs, err := storage.FindListByUserID(ctx, tt.userID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				if len(tt.wantURLs) > 0 {
					assert.ElementsMatch(t, tt.wantURLs, foundURLs, "Expected the URLs to match the stored values")
				} else {
					assert.Empty(t, foundURLs, "Expected no URLs to be found")
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
				{
					ShortID:     "abc123",
					OriginalURL: "http://example.com",
					UserID:      "user1",
					CreatedAt:   time.Time{},
				},
				{
					ShortID:     "xyz789",
					OriginalURL: "http://another-example.com",
					UserID:      "user2",
					CreatedAt:   time.Time{},
				},
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

	if _, err := storage.Insert(ctx, &model.URL{ShortID: "abc123", OriginalURL: "http://example.com", UserID: "user1"}); err != nil {
		t.Errorf("Insert() returned an error: %v", err)
	}
	if _, err := storage.Insert(ctx, &model.URL{ShortID: "xyz789", OriginalURL: "http://another-example.com", UserID: "user2"}); err != nil {
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
