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

	mockDB.ExpectExec(`CREATE TABLE IF NOT EXISTS urls`).
		WillReturnResult(pgxmock.NewResult("CREATE", 1))

	repo := database.NewURLRepository(mockDB)

	return repo, mockDB
}

// func TestURLRepository_Insert(t *testing.T) {
// 	ctx := context.Background()
// 	repo, mockDB := setupMockRepository(t)
// 	defer mockDB.Close()

// 	url := &model.URL{
// 		ShortID:     "abc123",
// 		OriginalURL: "http://example.com",
// 	}

// 	mockDB.ExpectQuery(`INSERT INTO urls`).
// 		WithArgs(url.ShortID, url.OriginalURL).
// 		WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "created_at"}).
// 			AddRow(1, url.ShortID, url.OriginalURL, time.Now()))

// 	insertedURL, err := repo.Insert(ctx, url)
// 	assert.NoError(t, err)
// 	assert.Equal(t, url.ShortID, insertedURL.ShortID)
// 	assert.Equal(t, url.OriginalURL, insertedURL.OriginalURL)
// }

func TestURLRepository_Insert(t *testing.T) {
	ctx := context.Background()
	repo, mockDB := setupMockRepository(t)
	defer mockDB.Close()

	tests := []struct {
		name          string
		shortID       string
		originalURL   string
		mockSetup     func()
		expectedError error
	}{
		{
			name:        "Successful Insert",
			shortID:     "abc123",
			originalURL: "http://example.com",
			mockSetup: func() {
				mockDB.ExpectQuery(`INSERT INTO urls`).
					WithArgs("abc123", "http://example.com").
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "created_at"}).
						AddRow(1, "abc123", "http://example.com", time.Now()))
			},
			expectedError: nil,
		},
		{
			name:        "Insert Error",
			shortID:     "abc123",
			originalURL: "http://example.com",
			mockSetup: func() {
				mockDB.ExpectQuery(`INSERT INTO urls`).
					WithArgs("abc123", "http://example.com").
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
				// Фиксируем текущее время в переменной для последующего сравнения
				fixedTime := time.Now()

				mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT id, short_id, original_url, created_at FROM urls WHERE short_id = $1`)).
					WithArgs("12345678").
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "created_at"}).
						AddRow(1, "12345678", "http://example.com", fixedTime))
			},
			expectedError: nil,
			expectedURL: &model.URL{
				ID:          1,
				ShortID:     "12345678",
				OriginalURL: "http://example.com",
			},
		},
		{
			name:    "FindByID Error - No Rows",
			shortID: "nonexistentID",
			mockSetup: func() {
				mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT id, short_id, original_url, created_at FROM urls WHERE short_id = $1`)).
					WithArgs("nonexistentID").
					WillReturnError(pgx.ErrNoRows) // Используем правильную ошибку из pgx
			},
			expectedError: pgx.ErrNoRows, // Ожидаем ошибку pgx.ErrNoRows
			expectedURL:   nil,           // Ожидаем, что результат будет nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			foundURL, err := repo.FindByID(ctx, tt.shortID)

			if tt.expectedError != nil {
				// Если ожидается ошибка, проверяем, что ошибка была и что она соответствует ожидаемой
				assert.Error(t, err)
				// Сравниваем ошибку напрямую
				assert.Equal(t, tt.expectedError, err)
			} else {
				// Если ошибки нет, проверяем, что возвращается правильный URL
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL.ShortID, foundURL.ShortID)
				assert.Equal(t, tt.expectedURL.OriginalURL, foundURL.OriginalURL)

				// Проверяем, что время совпадает с фиксированным временем (с небольшим допуском)
				assert.WithinDuration(t, foundURL.CreatedAt, time.Now(), 1*time.Second)
			}
		})
	}
}

func TestURLRepository_List(t *testing.T) {
	ctx := context.Background()
	repo, mockDB := setupMockRepository(t)
	defer mockDB.Close()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedURLs  []*model.URL
		expectedError error
	}{
		{
			name: "Successful List",
			mockSetup: func() {
				mockDB.ExpectQuery(`SELECT id, short_id, original_url, created_at FROM urls`).
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "created_at"}).
						AddRow(1, "abc123", "http://example.com", time.Now()).
						AddRow(2, "xyz789", "http://another-example.com", time.Now()))
			},
			expectedURLs: []*model.URL{
				{ID: 1, ShortID: "abc123", OriginalURL: "http://example.com", CreatedAt: time.Now()},
				{ID: 2, ShortID: "xyz789", OriginalURL: "http://another-example.com", CreatedAt: time.Now()},
			},
			expectedError: nil,
		},
		{
			name: "List Error",
			mockSetup: func() {
				mockDB.ExpectQuery(`SELECT id, short_id, original_url, created_at FROM urls`).
					WillReturnError(fmt.Errorf("error fetching URLs"))
			},
			expectedURLs:  nil,
			expectedError: fmt.Errorf("failed to find URLs: error fetching URLs"),
		},
		{
			name: "List Scan Error",
			mockSetup: func() {
				mockDB.ExpectQuery(`SELECT id, short_id, original_url, created_at FROM urls`).
					WillReturnRows(pgxmock.NewRows([]string{"id", "short_id", "original_url", "created_at"}).
						AddRow(1, "abc123", "http://example.com", time.Now()).
						AddRow(2, "xyz789", "http://another-example.com", time.Now()).
						RowError(1, fmt.Errorf("failed to scan URL data"))) // Симулируем ошибку на втором ряду
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
