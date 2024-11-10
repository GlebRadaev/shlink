package interfaces

import (
	"context"

	"github.com/GlebRadaev/shlink/internal/model"
)

type IURLRepository interface {
	Insert(ctx context.Context, url *model.URL) (*model.URL, error)
	FindByID(ctx context.Context, shortID string) (*model.URL, error)
	List(ctx context.Context) ([]*model.URL, error)
}
