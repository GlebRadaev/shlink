package health

import (
	"context"
	"errors"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/logger"
	"go.uber.org/zap"
)

type HealthService struct {
	log     *zap.SugaredLogger
	cfg     *config.Config
	urlRepo interfaces.IURLRepository
}

func NewHealthService(cfg *config.Config, log *logger.Logger, urlRepo interfaces.IURLRepository) *HealthService {
	return &HealthService{
		log:     log.Named("HealthService"),
		cfg:     cfg,
		urlRepo: urlRepo,
	}
}

func (s *HealthService) CheckDatabaseConnection(ctx context.Context) error {
	if err := s.urlRepo.Ping(ctx); err != nil {
		s.log.Error("Database connection error:", err)
		return errors.New("database connection error")
	}
	s.log.Info("Database connection is healthy.")
	return nil
}
