package database_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/jackc/pgx/v5"

	"github.com/GlebRadaev/shlink/internal/repository/database"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func setupMockRepository(t *testing.T) (interfaces.IURLRepository, pgxmock.PgxPoolIface) {
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	repo := database.NewURLRepository(mockDB)

	return repo, mockDB
}

func TestURLRepository_Insert(t *testing.T) {
	ctx := context.Background()
	repo, mockDB := setupMockRepository(t)
	defer mockDB.Close()

	tests := []struct {
		name          string
		shortID       string
		originalURL   string
		userID        string
		mockSetup     func()
		expectedError error
	}{
		{
			name:        "Successful Insert",
			shortID:     "abc123",
			originalURL: "http://example1.com",
			userID:      "user123",
			mockSetup: func() {
				mockDB.ExpectQuery(`INSERT INTO urls`).
					WithArgs("abc123", "http://example1.com", "user123").
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "user_id", "created_at"}).
						AddRow(1, "abc123", "http://example1.com", "user123", time.Now()))
			},
			expectedError: nil,
		},
		{
			name:        "Insert Error",
			shortID:     "abc123",
			originalURL: "http://example2.com",
			userID:      "user124",
			mockSetup: func() {
				mockDB.ExpectQuery(`INSERT INTO urls`).
					WithArgs("abc123", "http://example2.com", "user124").
					WillReturnError(errors.New("insert error"))
			},
			expectedError: errors.New("failed to insert URL: insert error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			url := &model.URL{
				ShortID:     tt.shortID,
				OriginalURL: tt.originalURL,
				UserID:      tt.userID,
			}

			insertedURL, err := repo.Insert(ctx, url)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.shortID, insertedURL.ShortID)
				assert.Equal(t, tt.originalURL, insertedURL.OriginalURL)
			}
		})
	}
}

func TestURLRepository_InsertList(t *testing.T) {
	ctx := context.Background()
	repo, mockDB := setupMockRepository(t)
	defer mockDB.Close()

	tests := []struct {
		name          string
		urls          []*model.URL
		mockSetup     func()
		expectedError error
	}{
		{
			name: "Successful InsertList",
			urls: []*model.URL{
				{ShortID: "abc123", OriginalURL: "http://example3.com", UserID: "user123"},
				{ShortID: "xyz789", OriginalURL: "http://another-example3.com", UserID: "user124"},
			},
			mockSetup: func() {
				mockDB.ExpectBegin()
				mockDB.ExpectQuery(`INSERT INTO urls`).
					WithArgs("abc123", "http://example3.com", "user123").
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "user_id", "created_at"}).
						AddRow(1, "abc123", "http://example3.com", "user123", time.Now()))
				mockDB.ExpectQuery(`INSERT INTO urls`).
					WithArgs("xyz789", "http://another-example3.com", "user124").
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "user_id", "created_at"}).
						AddRow(2, "xyz789", "http://another-example3.com", "user124", time.Now()))
				mockDB.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name: "InsertList Transaction Error",
			urls: []*model.URL{
				{ShortID: "abc123", OriginalURL: "http://example4.com", UserID: "user125"},
			},
			mockSetup: func() {
				mockDB.ExpectBegin().
					WillReturnError(fmt.Errorf("failed to begin transaction"))
			},
			expectedError: fmt.Errorf("failed to begin transaction: failed to begin transaction"),
		},
		{
			name: "InsertList Insert Error",
			urls: []*model.URL{
				{ShortID: "abc123", OriginalURL: "http://example4.com", UserID: "user125"},
				{ShortID: "xyz789", OriginalURL: "http://another-example4.com", UserID: "user126"},
			},
			mockSetup: func() {
				mockDB.ExpectBegin()
				mockDB.ExpectQuery(`INSERT INTO urls`).
					WithArgs("abc123", "http://example4.com", "user125").
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "user_id", "created_at"}).
						AddRow(1, "abc123", "http://example4.com", "user125", time.Now()))
				mockDB.ExpectQuery(`INSERT INTO urls`).
					WithArgs("xyz789", "http://another-example4.com", "user126").
					WillReturnError(fmt.Errorf("insert error"))
				mockDB.ExpectRollback()
			},
			expectedError: fmt.Errorf("failed to insert URL: insert error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			urls, err := repo.InsertList(ctx, tt.urls)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, urls, len(tt.urls))
				for i, url := range urls {
					assert.Equal(t, tt.urls[i].ShortID, url.ShortID)
					assert.Equal(t, tt.urls[i].OriginalURL, url.OriginalURL)
				}
			}
		})
	}
}

func TestURLRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	repo, mockDB := setupMockRepository(t)
	defer mockDB.Close()

	tests := []struct {
		name          string
		shortID       string
		mockSetup     func()
		expectedError error
		expectedURL   *model.URL
	}{
		{
			name:    "Successful FindByID",
			shortID: "12345678",
			mockSetup: func() {
				fixedTime := time.Now()

				mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT id, short_id, original_url, user_id, created_at FROM urls WHERE short_id = $1`)).
					WithArgs("12345678").
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "user_id", "created_at"}).
						AddRow(1, "12345678", "http://example.com", "user123", fixedTime))
			},
			expectedError: nil,
			expectedURL: &model.URL{
				ID:          1,
				ShortID:     "12345678",
				OriginalURL: "http://example.com",
				UserID:      "user123",
			},
		},
		{
			name:    "FindByID Error - No Rows",
			shortID: "nonexistentID",
			mockSetup: func() {
				mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT id, short_id, original_url, user_id, created_at FROM urls WHERE short_id = $1`)).
					WithArgs("nonexistentID").
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: pgx.ErrNoRows,
			expectedURL:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			foundURL, err := repo.FindByID(ctx, tt.shortID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL.ShortID, foundURL.ShortID)
				assert.Equal(t, tt.expectedURL.OriginalURL, foundURL.OriginalURL)
				assert.WithinDuration(t, foundURL.CreatedAt, time.Now(), 1*time.Second)
			}
		})
	}
}

func TestURLRepository_List(t *testing.T) {
	ctx := context.Background()
	repo, mockDB := setupMockRepository(t)
	defer mockDB.Close()
	fixedTime := time.Date(2024, time.November, 17, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		mockSetup     func()
		expectedURLs  []*model.URL
		expectedError error
	}{
		{
			name: "Successful List",
			mockSetup: func() {
				mockDB.ExpectQuery(`SELECT id, short_id, original_url, created_at, user_id FROM urls`).
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "created_at", "user_id"}).
						AddRow(1, "abc123", "http://example.com", fixedTime, "user123").
						AddRow(2, "xyz789", "http://another-example.com", fixedTime, "user456"))
			},
			expectedURLs: []*model.URL{
				{ID: 1, ShortID: "abc123", OriginalURL: "http://example.com", CreatedAt: fixedTime, UserID: "user123"},
				{ID: 2, ShortID: "xyz789", OriginalURL: "http://another-example.com", CreatedAt: fixedTime, UserID: "user456"},
			},
			expectedError: nil,
		},
		{
			name: "List Error",
			mockSetup: func() {
				mockDB.ExpectQuery(`SELECT id, short_id, original_url, created_at, user_id FROM urls`).
					WillReturnError(fmt.Errorf("error fetching URLs"))
			},
			expectedURLs:  nil,
			expectedError: fmt.Errorf("failed to find URLs: error fetching URLs"),
		},
		{
			name: "List Scan Error",
			mockSetup: func() {
				mockDB.ExpectQuery(`SELECT id, short_id, original_url, created_at, user_id FROM urls`).
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "created_at", "user_id"}).
						AddRow(1, "abc123", "http://example.com", fixedTime, "user123").
						AddRow(2, "xyz789", "http://another-example.com", fixedTime, "user456").
						RowError(1, fmt.Errorf("failed to scan URL data")))
			},
			expectedURLs:  nil,
			expectedError: fmt.Errorf("failed to scan URL data: failed to scan URL data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			urls, err := repo.List(ctx)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, urls, len(tt.expectedURLs))
				for i, url := range urls {
					assert.Equal(t, tt.expectedURLs[i].ShortID, url.ShortID)
					assert.Equal(t, tt.expectedURLs[i].OriginalURL, url.OriginalURL)
				}
			}
		})
	}
}

func TestURLRepository_Ping(t *testing.T) {
	ctx := context.Background()
	repo, mockDB := setupMockRepository(t)
	defer mockDB.Close()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedError error
	}{
		{
			name: "Successful Ping",
			mockSetup: func() {
				// Mock a successful database query for ping
				mockDB.ExpectQuery(`SELECT 1`).
					WillReturnRows(pgxmock.NewRows([]string{"result"}).AddRow(1))
			},
			expectedError: nil,
		},
		{
			name: "Ping Error",
			mockSetup: func() {
				// Simulate an error (e.g., database connection issue)
				mockDB.ExpectQuery(`SELECT 1`).
					WillReturnError(fmt.Errorf("db connection error"))
			},
			expectedError: fmt.Errorf("db connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := repo.Ping(ctx)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestURLRepository_ListByUserID(t *testing.T) {
	ctx := context.Background()
	repo, mockDB := setupMockRepository(t)
	defer mockDB.Close()

	tests := []struct {
		name          string
		mockSetup     func()
		userID        string
		expectedURLs  []*model.URL
		expectedError error
	}{
		{
			name: "Successful List By UserID",
			mockSetup: func() {
				mockDB.ExpectQuery(`SELECT id, short_id, original_url, user_id, created_at FROM urls WHERE user_id = \$1`).
					WithArgs("user123").
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "user_id", "created_at"}).
						AddRow(1, "abc123", "http://example.com", "user123", time.Now()))
			},
			userID: "user123",
			expectedURLs: []*model.URL{
				{ID: 1, ShortID: "abc123", OriginalURL: "http://example.com", CreatedAt: time.Now(), UserID: "user123"},
			},
			expectedError: nil,
		},
		{
			name: "List By UserID Error",
			mockSetup: func() {
				mockDB.ExpectQuery(`SELECT id, short_id, original_url, user_id, created_at FROM urls WHERE user_id = \$1`).
					WithArgs("user123").
					WillReturnError(fmt.Errorf("error fetching URLs for user 123"))
			},
			userID:        "user123",
			expectedURLs:  nil,
			expectedError: fmt.Errorf("error fetching URLs for user 123"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			urls, err := repo.FindListByUserID(ctx, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, urls, len(tt.expectedURLs))
				for i, url := range urls {
					assert.Equal(t, tt.expectedURLs[i].ShortID, url.ShortID)
					assert.Equal(t, tt.expectedURLs[i].OriginalURL, url.OriginalURL)
					assert.Equal(t, tt.expectedURLs[i].UserID, url.UserID)
				}
			}
		})
	}
}
