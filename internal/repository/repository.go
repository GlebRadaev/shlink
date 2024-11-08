package repository

import (
	"context"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/repository/database"
	"github.com/GlebRadaev/shlink/internal/repository/inmemory"
)

type Repositories struct {
	URLRepo interfaces.IURLRepository
}

func NewRepositoryFactory(ctx context.Context, cfg *config.Config) *Repositories {
	var urlRepo interfaces.IURLRepository
	if cfg.DatabaseDSN != "" {
		urlRepo, _ = database.NewURLRepository(ctx, cfg.DatabaseDSN)
	} else {
		urlRepo = inmemory.NewMemoryStorage()
	}

	return &Repositories{
		URLRepo: urlRepo,
	}
}
