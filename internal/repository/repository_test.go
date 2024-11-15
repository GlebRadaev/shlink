package repository

import (
	"context"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository/database"
	"github.com/GlebRadaev/shlink/internal/repository/inmemory"
	"github.com/stretchr/testify/assert"
)

func TestNewRepositoryFactory(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.NewLogger("info")

	t.Run("creates database URLRepository if DSN is provided", func(t *testing.T) {
		cfg := &config.Config{DatabaseDSN: "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"}
		repos := NewRepositoryFactory(ctx, cfg, log)

		_, ok := repos.URLRepo.(*database.URLRepository)
		assert.True(t, ok, "Expected a URLRepository instance")

		_, isMemoryRepo := repos.URLRepo.(*inmemory.MemoryStorage)
		assert.False(t, isMemoryRepo, "Expected not to be a MemoryStorage instance")
	})

	t.Run("creates in-memory MemoryStorage if DSN is invalid", func(t *testing.T) {
		cfg := &config.Config{DatabaseDSN: "invalid-dsn"}
		repos := NewRepositoryFactory(ctx, cfg, log)

		_, ok := repos.URLRepo.(*inmemory.MemoryStorage)
		assert.True(t, ok, "Expected a MemoryStorage instance")

		_, isDatabaseRepo := repos.URLRepo.(*database.URLRepository)
		assert.False(t, isDatabaseRepo, "Expected not to be a database repository")
	})

	t.Run("creates in-memory MemoryStorage if DSN is empty", func(t *testing.T) {
		cfg := &config.Config{DatabaseDSN: ""}
		repos := NewRepositoryFactory(ctx, cfg, log)

		_, ok := repos.URLRepo.(*inmemory.MemoryStorage)
		assert.True(t, ok, "Expected a MemoryStorage instance")

		_, isDatabaseRepo := repos.URLRepo.(*database.URLRepository)
		assert.False(t, isDatabaseRepo, "Expected not to be a database repository")
	})
}
