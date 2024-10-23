package repository

import (
	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/repository/filestorage"
	"github.com/GlebRadaev/shlink/internal/repository/inmemory"
)

func RepositoryFactory(config *config.Config) (interfaces.Repository, interfaces.Repository) {
	memoryRepo := inmemory.NewMemoryStorage()
	fileRepo := filestorage.NewFileStorage(config.FileStoragePath)
	return memoryRepo, fileRepo
}
