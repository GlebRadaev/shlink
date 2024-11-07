package interfaces

import "context"

type Repository interface {
	AddURL(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, bool, error)
	GetAll(ctx context.Context) (map[string]string, error)
}
