package service

import (
	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service/backup"
	"github.com/GlebRadaev/shlink/internal/service/url"
)

type Services struct {
	URLService    *url.URLService
	BackupService *backup.BackupService
}

type URLService = url.URLService

func NewServiceFactory(cfg *config.Config, log *logger.Logger, repos *repository.Repositories) *Services {
	backupService := backup.NewBackupService(cfg.FileStoragePath)
	urlService := url.NewURLService(cfg, log, backupService, repos.MemoryRepo)
	if err := urlService.LoadData(); err != nil {
		log.Errorf("Failed to load data: %v", err)
	}
	return &Services{
		URLService:    urlService,
		BackupService: backupService,
	}
}
