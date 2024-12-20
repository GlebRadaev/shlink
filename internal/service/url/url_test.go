package url_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service/backup"
	"github.com/GlebRadaev/shlink/internal/service/url"
	"github.com/GlebRadaev/shlink/internal/utils"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

var cfg *config.Config

func setup(t *testing.T) (*repository.MockIURLRepository, *url.URLService, *backup.MockIBackupService, *config.Config, error) {
	if cfg == nil {
		var err error
		cfg, err = config.ParseAndLoadConfig()
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}
	log, _ := logger.NewLogger("info")
	ctrl := gomock.NewController(t)
	mockURLRepo := repository.NewMockIURLRepository(ctrl)
	mockBackupService := backup.NewMockIBackupService(ctrl)
	urlService := url.NewURLService(cfg, log, mockBackupService, mockURLRepo)
	defer ctrl.Finish()

	return mockURLRepo, urlService, mockBackupService, cfg, nil
}

func TestURLService_LoadData(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, urlService, mockBackupService, _, err := setup(t)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "LoadData success",
			setupMock: func() {
				mockData := map[string]string{"testID": "http://example1.com"}
				mockBackupService.EXPECT().LoadData().Return(mockData, nil)
				mockURLRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(&model.URL{ShortID: "testID", OriginalURL: "http://example1.com"}, nil)
			},
			wantErr: nil,
		},
		{
			name: "LoadData error",
			setupMock: func() {
				mockBackupService.EXPECT().LoadData().Return(nil, errors.New("load error"))
			},
			wantErr: errors.New("load error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := urlService.LoadData(ctx)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestURLService_SaveData(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, urlService, mockBackupService, _, err := setup(t)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "SaveData success",
			setupMock: func() {
				mockURLRepo.EXPECT().List(gomock.Any()).Return([]*model.URL{{ShortID: "testID", OriginalURL: "http://example2.com"}}, nil)
				mockBackupService.EXPECT().SaveData(gomock.Any()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "SaveData error",
			setupMock: func() {
				mockURLRepo.EXPECT().List(gomock.Any()).Return(nil, errors.New("list error"))
			},
			wantErr: errors.New("list error"),
		},
		{
			name: "SaveData backup service error",
			setupMock: func() {
				mockURLRepo.EXPECT().List(gomock.Any()).Return([]*model.URL{{ShortID: "testID", OriginalURL: "http://example2.com"}}, nil)
				mockBackupService.EXPECT().SaveData(gomock.Any()).Return(errors.New("save data error"))
			},
			wantErr: errors.New("save data error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := urlService.SaveData(ctx)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestURLService_Shorten(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, urlService, _, cfg, err := setup(t)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}
	type args struct {
		url string
	}
	tests := []struct {
		name      string
		setupMock func(mockURLRepo *repository.MockIURLRepository)
		args      args
		wantErr   error
	}{
		// {
		// 	name: "valid http URL",
		// 	setupMock: func(mockURLRepo *repository.MockIURLRepository) {
		// 		mockURLRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(&model.URL{ShortID: "testID12", OriginalURL: "http://example3.com"}, nil)
		// 		mockURLRepo.EXPECT().FindByID(ctx, "testID12").Return(&model.URL{ShortID: "testID12", OriginalURL: "http://example3.com"}, nil)
		// 	},
		// 	args:    args{url: "http://example3.com"},
		// 	wantErr: nil,
		// },
		// {
		// 	name: "valid https URL",
		// 	args: args{url: "https://example4.com"},
		// 	setupMock: func(mockURLRepo *repository.MockIURLRepository) {
		// 		mockURLRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(&model.URL{ShortID: "testID12", OriginalURL: "https://example4.com"}, nil)
		// 		mockURLRepo.EXPECT().FindByID(ctx, "testID12").Return(&model.URL{ShortID: "testID12", OriginalURL: "https://example4.com"}, nil)
		// 	},
		// 	wantErr: nil,
		// },
		{
			name: "error inserting URL into memory repository",
			args: args{url: "https://example.com"},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil, errors.New("error inserting URL"))
			},
			wantErr: errors.New("error inserting URL"),
		},
		{
			name: "invalid URL scheme",
			args: args{url: "httpsss://example.com"},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil, errors.New("invalid URL scheme"))
			},
			wantErr: errors.New("invalid URL scheme"),
		},
		{
			name: "invalid URL format (URL too long)",
			args: args{url: string(make([]byte, 2048+1))},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil, errors.New("invalid URL format"))
			},
			wantErr: errors.New("invalid URL format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockURLRepo)
			got, err := urlService.Shorten(ctx, tt.args.url)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				expectedLength := len(cfg.BaseURL) + 1 + 8
				assert.Equal(t, expectedLength, len(got), "Expected ID length to be %d, but got %d", expectedLength, len(got))
				shortID := strings.Split(got, "/")[len(strings.Split(got, "/"))-1]
				storedURL, _ := mockURLRepo.FindByID(ctx, shortID)
				assert.Equal(t, tt.args.url, storedURL.OriginalURL, "Stored URL mismatch in memoryRepo: got %v, want %v", storedURL.OriginalURL, tt.args.url)
			}
		})
	}
}

func TestURLService_ShortenList(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, urlService, _, cfg, err := setup(t)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	tests := []struct {
		name        string
		data        dto.BatchShortenRequestDTO
		setupMock   func(mockURLRepo *repository.MockIURLRepository)
		wantErr     error
		expectedLen int
	}{
		{
			name: "Batch shorten success",
			data: dto.BatchShortenRequestDTO{
				{CorrelationID: "1", OriginalURL: "http://example1.com"},
				{CorrelationID: "2", OriginalURL: "https://example2.com"},
			},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().InsertList(ctx, gomock.Any()).Return([]*model.URL{
					{ShortID: "short1", OriginalURL: "http://example1.com"},
					{ShortID: "short2", OriginalURL: "https://example2.com"},
				}, nil)
			},
			wantErr:     nil,
			expectedLen: 2,
		},
		{
			name: "Batch shorten with invalid URL",
			data: dto.BatchShortenRequestDTO{
				{CorrelationID: "1", OriginalURL: "invalid-url"},
				{CorrelationID: "2", OriginalURL: "https://example2.com"},
			},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().InsertList(ctx, gomock.Any()).Return([]*model.URL{
					{ShortID: "short2", OriginalURL: "https://example2.com"},
				}, nil)
			},
			wantErr:     nil,
			expectedLen: 1,
		},
		{
			name: "Batch shorten repository error",
			data: dto.BatchShortenRequestDTO{
				{CorrelationID: "1", OriginalURL: "http://example1.com"},
			},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().InsertList(ctx, gomock.Any()).Return(nil, errors.New("repository error"))
			},
			wantErr:     errors.New("repository error"),
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockURLRepo)
			got, err := urlService.ShortenList(ctx, tt.data)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				assert.Nil(t, got, "Expected no result when there's an error")
			} else {
				assert.NoError(t, err, "Unexpected error in ShortenList")

				assert.Equal(t, tt.expectedLen, len(got), "Unexpected number of successful URLs")

				for i, res := range got {
					assert.True(t, strings.HasPrefix(res.ShortURL, cfg.BaseURL), "ShortURL should start with base URL")
					if i < len(tt.data) && utils.IsValidID(res.ShortURL, 8) {
						assert.Equal(t, tt.data[i].CorrelationID, res.CorrelationID, "CorrelationID should match")
					}
				}
			}
		})
	}
}

func TestURLService_GetOriginal(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, urlService, _, _, err := setup(t)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}
	type args struct {
		id string
	}
	tests := []struct {
		name      string
		args      args
		setupMock func(mockURLRepo *repository.MockIURLRepository)
		want      string
		wantErr   error
	}{
		{
			name: "valid ID",
			args: args{id: "testIDid"},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().FindByID(gomock.Any(), gomock.Eq("testIDid")).Return(&model.URL{ShortID: "testIDid", OriginalURL: "http://example.com"}, nil)
			},
			want:    "http://example.com",
			wantErr: nil,
		},
		{
			name:      "invalid ID length",
			args:      args{id: "short"},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {},
			want:      "",
			wantErr:   errors.New("invalid ID"),
		},
		{
			name:      "invalid ID format",
			args:      args{id: "invalid!"},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {},
			want:      "",
			wantErr:   errors.New("invalid ID"),
		},
		{
			name: "ID not found",
			args: args{id: "notfound"},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().FindByID(gomock.Any(), gomock.Eq("notfound")).Return(nil, errors.New("URL not found"))
			},
			want:    "",
			wantErr: errors.New("URL not found"),
		},
		{
			name: "ID not found (nil URL)",
			args: args{id: "notfound"},
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().FindByID(gomock.Any(), gomock.Eq("notfound")).Return(nil, nil)
			},
			want:    "",
			wantErr: errors.New("URL not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockURLRepo)
			got, err := urlService.GetOriginal(ctx, tt.args.id)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
