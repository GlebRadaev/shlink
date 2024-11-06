package repository

import (
	"context"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/repository/database"
	"github.com/GlebRadaev/shlink/internal/repository/inmemory"
)

type Repositories struct {
	MemoryRepo   interfaces.Repository
	DatabaseRepo *database.DatabaseRepository
}

func NewRepositoryFactory(ctx context.Context, cfg *config.Config) *Repositories {
	memoryRepo := inmemory.NewMemoryStorage()
	databaseRepo := database.NewDatabaseRepository(ctx, cfg)

	return &Repositories{
		MemoryRepo:   memoryRepo,
		DatabaseRepo: databaseRepo,
	}
}
