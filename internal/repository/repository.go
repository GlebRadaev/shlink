package repository

import (
	"context"
	"database/sql"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository/database"
	"github.com/GlebRadaev/shlink/internal/repository/inmemory"
	"github.com/GlebRadaev/shlink/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Repositories struct {
	URLRepo interfaces.IURLRepository
}

func NewRepositoryFactory(ctx context.Context, cfg *config.Config, log *logger.Logger) *Repositories {
	var urlRepo interfaces.IURLRepository
	logger := log.Named("RepositoryFactory")
	if cfg.DatabaseDSN != "" {
		pool, err := pgxpool.New(ctx, cfg.DatabaseDSN)
		if err == nil {
			logger.Info("Connected to database.")
			if err := Migrate(ctx, cfg.DatabaseDSN); err != nil {
				logger.Error("Failed to run migrations: %v", err)
			}
			urlRepo = database.NewURLRepository(pool)
		} else {
			logger.Info("Connected to in-memory storage (failed to connect to database): %v", err)
			urlRepo = inmemory.NewMemoryStorage()
		}
	} else {
		logger.Info("Connected to in-memory storage.")
		urlRepo = inmemory.NewMemoryStorage()
	}

	return &Repositories{URLRepo: urlRepo}
}

func Migrate(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations.Migrations)
	return goose.RunContext(ctx, "up", db, ".")
}
