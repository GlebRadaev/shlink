package service

import (
	"context"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service/backup"
	"github.com/GlebRadaev/shlink/internal/service/health"
	"github.com/GlebRadaev/shlink/internal/service/url"
)

type Services struct {
	URLService    *url.URLService
	BackupService *backup.BackupService
	HealthService *health.HealthService
}

type URLService = url.URLService
type HealthService = health.HealthService

func NewServiceFactory(ctx context.Context, cfg *config.Config, log *logger.Logger, repos *repository.Repositories) *Services {
	logger := log.Named("ServiceFactory")
	backupService := backup.NewBackupService(cfg.FileStoragePath)
	logger.Info("Backup service up.")
	urlService := url.NewURLService(cfg, log, backupService, repos.URLRepo)
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
