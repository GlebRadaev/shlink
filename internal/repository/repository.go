package repository

import (
	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/repository/inmemory"
)

type Repositories struct {
	MemoryRepo interfaces.Repository
}

func NewRepositoryFactory(cfg *config.Config) *Repositories {
	memoryRepo := inmemory.NewMemoryStorage()

	return &Repositories{
		MemoryRepo: memoryRepo,
	}
}
