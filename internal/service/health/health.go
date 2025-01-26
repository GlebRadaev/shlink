// Package health provides services for checking the health of the application.
// It includes a HealthService that can check the database connection and report its status.
package health

import (
	"context"
	"errors"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/logger"
	"go.uber.org/zap"
)

// HealthService provides functionality to check the health of the application
// by performing checks like database connection status.
type HealthService struct {
	// log is the logger used for logging health-related information.
	log *zap.SugaredLogger

	// cfg holds the application configuration.
	cfg *config.Config

	// urlRepo is the URL repository used to check the database connection.
	urlRepo interfaces.IURLRepository
}

// NewHealthService creates and returns a new instance of HealthService.
func NewHealthService(cfg *config.Config, log *logger.Logger, urlRepo interfaces.IURLRepository) *HealthService {
	return &HealthService{
		log:     log.Named("HealthService"),
		cfg:     cfg,
		urlRepo: urlRepo,
	}
}

// CheckDatabaseConnection checks the connection to the database by pinging the URL repository.
func (s *HealthService) CheckDatabaseConnection(ctx context.Context) error {
	// Attempt to ping the database
	if err := s.urlRepo.Ping(ctx); err != nil {
		// If the ping fails, log and return an error
		s.log.Error("Database connection error:", err)
		return errors.New("database connection error")
	}
	// If the connection is healthy, log and return nil
	s.log.Info("Database connection is healthy.")
	return nil
}
