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
	backupService := backup.NewBackupService(cfg.FileStoragePath)
	urlService := url.NewURLService(cfg, log, backupService, repos.URLRepo)
	healthService := health.NewHealthService(cfg, log, repos.URLRepo)

	if cfg.FileStoragePath != "" {
		if err := urlService.LoadData(ctx); err != nil {
			log.Errorf("Failed to load data: %v", err)
		}
	}
	return &Services{
		URLService:    urlService,
		BackupService: backupService,
		HealthService: healthService,
	}
}
