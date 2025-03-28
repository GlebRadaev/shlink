package url_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/dto"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service/backup"
	"github.com/GlebRadaev/shlink/internal/service/url"
	"github.com/GlebRadaev/shlink/internal/taskmanager"
	"github.com/GlebRadaev/shlink/internal/utils"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cfg *config.Config

func setup(t *testing.T, ctx context.Context) (*repository.MockIURLRepository, *url.URLService, *backup.MockIBackupService, *config.Config, *taskmanager.WorkerPool, error) {
	if cfg == nil {
		var err error
		cfg, err = config.ParseAndLoadConfig()
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}
	log, _ := logger.NewLogger("info")
	ctrl := gomock.NewController(t)
	pool := taskmanager.NewWorkerPool(ctx, 10, 1)
	defer pool.Shutdown()
	mockURLRepo := repository.NewMockIURLRepository(ctrl)
	mockBackupService := backup.NewMockIBackupService(ctrl)
	urlService := url.NewURLService(cfg, log, pool, mockBackupService, mockURLRepo)
	defer ctrl.Finish()

	return mockURLRepo, urlService, mockBackupService, cfg, pool, nil
}

func TestURLService_LoadData(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, urlService, mockBackupService, _, _, err := setup(t, ctx)
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
	mockURLRepo, urlService, mockBackupService, _, _, err := setup(t, ctx)
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

func TestURLService_processDeleteURLsTask(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, urlService, _, _, _, err := setup(t, ctx)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	testCases := []struct {
		name        string
		task        taskmanager.Task
		setupMock   func(mockURLRepo *repository.MockIURLRepository)
		expectedErr string
	}{
		{
			name: "successfully processes delete URLs task",
			task: taskmanager.DeleteTask{
				UserID: "user-123",
				URLs:   []string{"short1", "short2", "short3"},
			},
			// mockSetup: func(repo *mocks.IURLRepository) {
			// 	repo.On("DeleteListByUserIDAndShortIDs", mock.Anything, "user-123", mock.Anything).Return(nil)
			// },
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().DeleteListByUserIDAndShortIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedErr: "",
		},
		// {
		// 	name: "invalid task type error",
		// 	task: InvalidTask{},
		// 	setupMock: func(mockURLRepo *repository.MockIURLRepository) {
		// 	},
		// 	expectedErr: "invalid task type",
		// },
		{
			name: "batch deletion error",
			task: taskmanager.DeleteTask{
				UserID: "user-123",
				URLs:   []string{"short1", "short2"},
			},

			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().DeleteListByUserIDAndShortIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			expectedErr: "db error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(mockURLRepo)
			err := urlService.ProcessDeleteURLsTask(ctx, tc.task)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestURLService_Shorten(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, urlService, _, cfg, _, err := setup(t, ctx)
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
			got, err := urlService.Shorten(ctx, "user123", tt.args.url)
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
	mockURLRepo, urlService, _, cfg, _, err := setup(t, ctx)
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
			got, err := urlService.ShortenList(ctx, "user123", tt.data)

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
	mockURLRepo, urlService, _, _, _, err := setup(t, ctx)
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

func TestURLService_GetUserURLs(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, urlService, _, cfg, _, err := setup(t, ctx)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	tests := []struct {
		name      string
		userID    string
		setupMock func(mockURLRepo *repository.MockIURLRepository)
		want      dto.GetUserURLsResponseDTO
		wantErr   error
	}{
		{
			name:   "success",
			userID: "testUser",
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().FindListByUserID(gomock.Any(), gomock.Eq("testUser")).Return([]*model.URL{
					{ShortID: "short1", OriginalURL: "http://example1.com", UserID: "testUser"},
					{ShortID: "short2", OriginalURL: "https://example2.com", UserID: "testUser"},
				}, nil)
			},
			want: dto.GetUserURLsResponseDTO{
				{ShortURL: fmt.Sprintf("%s/short1", cfg.BaseURL), OriginalURL: "http://example1.com"},
				{ShortURL: fmt.Sprintf("%s/short2", cfg.BaseURL), OriginalURL: "https://example2.com"},
			},
			wantErr: nil,
		},
		{
			name:   "not found",
			userID: "notFoundUser",
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().FindListByUserID(gomock.Any(), gomock.Eq("notFoundUser")).Return([]*model.URL{}, nil)
			},
			want:    dto.GetUserURLsResponseDTO{},
			wantErr: nil,
		},
		{
			name:   "repository error",
			userID: "errorUser",
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().FindListByUserID(gomock.Any(), gomock.Eq("errorUser")).Return(nil, errors.New("repository error"))
			},
			want:    nil,
			wantErr: errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockURLRepo)
			got, err := urlService.GetUserURLs(ctx, tt.userID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestURLService_IsAllowed(t *testing.T) {
	ctx := context.Background()

	_, urlService, _, cfg, _, err := setup(t, ctx)

	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	cfg.TrustedSubnet = subnet.String()

	fmt.Print(cfg)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "IP outside trusted subnet",
			ip:       "10.0.0.1",
			expected: false,
		},
		{
			name:     "Invalid IP",
			ip:       "invalid-ip",
			expected: false,
		},
		{
			name:     "Empty IP",
			ip:       "",
			expected: false,
		},
		{
			name:     "No trusted subnet",
			ip:       "192.168.1.100",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "No trusted subnet" {
				cfg.TrustedSubnet = ""
			}
			result := urlService.IsAllowed(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestURLService_GetStats(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, urlService, _, _, _, err := setup(t, ctx)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	tests := []struct {
		name      string
		setupMock func(mockURLRepo *repository.MockIURLRepository)
		want      map[string]int
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().CountURLs(ctx).Return(10, nil)
				mockURLRepo.EXPECT().CountUsers(ctx).Return(5, nil)
			},
			want: map[string]int{
				"urls":  10,
				"users": 5,
			},
			wantErr: nil,
		},
		{
			name: "error counting URLs",
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().CountURLs(ctx).Return(0, errors.New("repository error"))
			},
			want:    nil,
			wantErr: errors.New("repository error"),
		},
		{
			name: "error counting users",
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().CountURLs(ctx).Return(10, nil)
				mockURLRepo.EXPECT().CountUsers(ctx).Return(0, errors.New("repository error"))
			},
			want:    nil,
			wantErr: errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockURLRepo)
			got, err := urlService.GetStats(ctx)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
