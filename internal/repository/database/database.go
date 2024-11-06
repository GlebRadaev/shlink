package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DatabaseRepository struct {
	db     *pgxpool.Pool
	cfg    *config.Config
	ctx    context.Context
	dbOnce sync.Once
}

func NewDatabaseRepository(ctx context.Context, cfg *config.Config) *DatabaseRepository {
	return &DatabaseRepository{
		cfg: cfg,
		ctx: ctx,
	}
}

func (r *DatabaseRepository) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

func (r *DatabaseRepository) initDB() error {
	var err error
	r.dbOnce.Do(func() {
		dbPool, poolErr := pgxpool.Connect(r.ctx, r.cfg.DatabaseDSN)
		if poolErr != nil {
			err = fmt.Errorf("failed to connect to database: %v", poolErr)
			return
		}
		r.db = dbPool
	})
	return err
}

func (r *DatabaseRepository) GetData(ctx context.Context, id int) (string, error) {
	if err := r.initDB(); err != nil {
		return "", fmt.Errorf("database connection error: %v", err)
	}
	return "sample data", nil
}

func (r *DatabaseRepository) Ping(ctx context.Context) error {
	if err := r.initDB(); err != nil {
		return err
	}
	if err := r.db.Ping(ctx); err != nil {
		return fmt.Errorf("database connection error: %v", err)
	}
	return nil
}
