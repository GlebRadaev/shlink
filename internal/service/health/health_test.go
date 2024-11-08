package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/service/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setup() (*health.HealthService, *MockURLRepository, *config.Config, error) {
	log, _ := logger.NewLogger("info")
	cfg := &config.Config{DatabaseDSN: ""}
	mockRepo := new(MockURLRepository)
	healthService := health.NewHealthService(cfg, log, mockRepo)
	return healthService, mockRepo, cfg, nil
}

type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) Insert(ctx context.Context, url *model.URL) (*model.URL, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*model.URL), args.Error(1)
}

func (m *MockURLRepository) FindByID(ctx context.Context, shortID string) (*model.URL, error) {
	args := m.Called(ctx, shortID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), nil
}

func (m *MockURLRepository) List(ctx context.Context) ([]*model.URL, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.URL), args.Error(1)
}

func TestHealthService_CheckDatabaseConnection(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		mockSetup func(mockRepo *MockURLRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "healthy connection",
			mockSetup: func(mockRepo *MockURLRepository) {
				mockRepo.On("FindByID", ctx, "checkQuery").Return(&model.URL{}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "database connection error",
			mockSetup: func(mockRepo *MockURLRepository) {
				mockRepo.On("FindByID", ctx, "checkQuery").Return(nil, errors.New("database connection error")).Once()
			},
			wantErr: true,
			errMsg:  "database connection error",
		},
	}

	healthService, mockRepo, _, err := setup()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mockRepo)
			err = healthService.CheckDatabaseConnection(ctx)
			if tt.wantErr {
				t.Logf("Expected error, got: %v", err)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				t.Logf("No error expected, got: %v", err)
				assert.NoError(t, err)
			}
			mockRepo.AssertCalled(t, "FindByID", ctx, "checkQuery")
			mockRepo.AssertExpectations(t)
		})
	}
}
