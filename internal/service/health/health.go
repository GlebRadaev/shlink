package health

import (
	"context"
	"errors"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository/database"
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
	if dbRepo, ok := s.urlRepo.(*database.URLRepository); ok {
		// Используем Ping, если это репозиторий с подключением к базе данных
		if err := dbRepo.Ping(ctx); err != nil {
			s.log.Error("Database connection error:", err)
			return errors.New("database connection error")
		}
	} else {
		// Если это in-memory хранилище вызываем FindByID
		if _, err := s.urlRepo.FindByID(ctx, "checkQuery"); err != nil {
			s.log.Error("Storage connection error:", err)
			return errors.New("storage connection error")
		}
	}
	s.log.Info("Database connection is healthy.")
	return nil
}
