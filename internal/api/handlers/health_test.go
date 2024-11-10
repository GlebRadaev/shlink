package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service/health"
	"go.uber.org/mock/gomock"
)

var cfgTest *config.Config

func setupHealth() (*logger.Logger, *config.Config, error) {
	if cfgTest == nil {
		var err error
		cfgTest, err = config.ParseAndLoadConfig()
		if err != nil {
			return nil, nil, err
		}
	}
	log, _ := logger.NewLogger("info")

	return log, cfgTest, nil
}

func TestHealthHandlers_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log, cfg, _ := setupHealth()
	mockRepo := repository.NewMockIURLRepository(ctrl)
	healthService := health.NewHealthService(cfg, log, mockRepo)
	handler := NewHealthHandlers(healthService)

	tests := []struct {
		name       string
		setup      func(*repository.MockIURLRepository)
		wantStatus int
	}{
		{
			name: "database connection is healthy",
			setup: func(repo *repository.MockIURLRepository) {
				repo.EXPECT().FindByID(gomock.Any(), "checkQuery").Return(&model.URL{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "database connection error",
			setup: func(repo *repository.MockIURLRepository) {
				repo.EXPECT().FindByID(gomock.Any(), "checkQuery").Return(nil, errors.New("database connection error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(mockRepo)
			req, err := http.NewRequest("GET", "/ping", nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			handler.Ping(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("expected status code %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
