package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/service/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) Insert(ctx context.Context, url *model.URL) (*model.URL, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*model.URL), args.Error(1)
}

func (m *MockURLRepository) FindByID(ctx context.Context, shortID string) (*model.URL, error) {
	args := m.Called(ctx, shortID)
	if args.Get(0) != nil {
		return args.Get(0).(*model.URL), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockURLRepository) List(ctx context.Context) ([]*model.URL, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.URL), args.Error(1)
}

var cfg *config.Config

func setupHealth() (*logger.Logger, *config.Config, error) {
	if cfg == nil {
		var err error
		cfg, err = config.ParseAndLoadConfig()
		if err != nil {
			return nil, nil, err
		}
	}
	log, _ := logger.NewLogger("info")

	return log, cfg, nil
}

func TestPingHandler(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*MockURLRepository) // Setup mock behavior
		wantStatus int
	}{
		{
			name: "database connection is healthy",
			setup: func(repo *MockURLRepository) {
				repo.On("FindByID", mock.Anything, "checkQuery").Return(&model.URL{}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "database connection error",
			setup: func(repo *MockURLRepository) {
				repo.On("FindByID", mock.Anything, "checkQuery").Return(nil, errors.New("database connection error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	// Настройка сервиса
	log, cfg, err := setupHealth()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	// Мокируем репозиторий
	mockURLRepo := new(MockURLRepository)

	// Создайте сервис, передав мокированный репозиторий
	healthService := health.NewHealthService(cfg, log, mockURLRepo)

	// Создайте обработчик, используя сервис
	handler := &HealthHandlers{
		healthService: healthService,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Info("Test: " + tt.name)
			tt.setup(mockURLRepo)

			req := httptest.NewRequest("GET", "/ping", nil)
			w := httptest.NewRecorder()

			handler.Ping(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			mockURLRepo.AssertExpectations(t)
		})
	}
}
