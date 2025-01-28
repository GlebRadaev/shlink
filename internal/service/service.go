// Package service provides the core services for URL management, backup, and health checks.
//
// This package includes the initialization of key services such as URLService, BackupService,
// and HealthService. It ties them together with their respective repositories and configuration
// to provide an abstraction layer for managing URLs, backups, and service health status in the system.
//
// Services:
// - URLService: Handles the business logic for URL shortening, management, and retrieval.
// - BackupService: Manages data backup and restoration, including saving and loading URL data.
// - HealthService: Provides health check endpoints for monitoring service status.
package service

import (
	"context"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service/backup"
	"github.com/GlebRadaev/shlink/internal/service/health"
	"github.com/GlebRadaev/shlink/internal/service/url"
	"github.com/GlebRadaev/shlink/internal/taskmanager"
)

// Services aggregates the primary services in the application: URLService, BackupService, and HealthService.
// It is used to interact with the core functionalities of URL shortening, backup management, and health checks.
type Services struct {
	URLService    *url.URLService       // Service for shortening URLs and managing URL data.
	BackupService *backup.BackupService // Service for performing data backup and restoration.
	HealthService *health.HealthService // Service for monitoring the application's health.
}

// URLService is an alias for url.URLService, providing the URL service functionalities.
type URLService = url.URLService

// HealthService is an alias for health.HealthService, providing health check functionalities.
type HealthService = health.HealthService

// NewServiceFactory initializes and returns an instance of Services, containing all core services
// needed to operate the system.
func NewServiceFactory(ctx context.Context, cfg *config.Config, log *logger.Logger, pool *taskmanager.WorkerPool, repos *repository.Repositories) *Services {
	logger := log.Named("ServiceFactory")

	backupService := backup.NewBackupService(cfg.FileStoragePath)
	logger.Info("Backup service up.")
	urlService := url.NewURLService(cfg, log, pool, backupService, repos.URLRepo)
	logger.Info("URL service up.")
	healthService := health.NewHealthService(cfg, log, repos.URLRepo)
	logger.Info("Health service up.")

	if err := urlService.LoadData(ctx); err != nil {
		logger.Errorf("Failed to load data: %v", err)
	} else {
		logger.Info("Data successfully loaded from backup.")
	}

	return &Services{
		URLService:    urlService,
		BackupService: backupService,
		HealthService: healthService,
	}
}
