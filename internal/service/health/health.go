package health

import (
	"context"
	"fmt"

	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository/database"
	"go.uber.org/zap"
)

type HealthService struct {
	log          *zap.SugaredLogger
	databaseRepo *database.DatabaseRepository
}

func NewHealthService(log *logger.Logger, databaseRepo *database.DatabaseRepository) *HealthService {
	return &HealthService{
		log:          log.Named("HealthService"),
		databaseRepo: databaseRepo,
	}
}

func (s *HealthService) CheckDatabaseConnection(ctx context.Context) error {
	if err := s.databaseRepo.Ping(ctx); err != nil {
		s.log.Error("Database connection error:", err)
		return fmt.Errorf("database connection error: %w", err)
	}
	s.log.Info("Database connection is healthy.")
	return nil
}
