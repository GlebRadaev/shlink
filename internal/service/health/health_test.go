package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service/health"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var cfg *config.Config

func setup(t *testing.T) (*repository.MockIURLRepository, *health.HealthService, *config.Config, error) {
	if cfg == nil {
		var err error
		cfg, err = config.ParseAndLoadConfig()
		if err != nil {
			return nil, nil, nil, err
		}
	}
	log, _ := logger.NewLogger("info")
	ctrl := gomock.NewController(t)
	mockURLRepo := repository.NewMockIURLRepository(ctrl)
	healthService := health.NewHealthService(cfg, log, mockURLRepo)
	defer ctrl.Finish()

	return mockURLRepo, healthService, cfg, nil
}

func TestHealthService_CheckDatabaseConnection(t *testing.T) {
	ctx := context.Background()
	mockURLRepo, healthService, _, err := setup(t)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	tests := []struct {
		name      string
		setupMock func(mockURLRepo *repository.MockIURLRepository)
		wantErr   error
	}{
		{
			name: "Database connection is healthy",
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().Ping(gomock.Any()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Database connection error",
			setupMock: func(mockURLRepo *repository.MockIURLRepository) {
				mockURLRepo.EXPECT().Ping(gomock.Any()).Return(errors.New("database connection error"))
			},
			wantErr: errors.New("database connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockURLRepo)
			err = healthService.CheckDatabaseConnection(ctx)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
