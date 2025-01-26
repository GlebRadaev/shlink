package interfaces

import (
	"context"

	"github.com/GlebRadaev/shlink/internal/model"
)

// IURLRepository defines the interface for URL-related data access operations.
type IURLRepository interface {
	// Insert adds a new URL entry to the repository.
	// Returns the created URL model or an error.
	Insert(ctx context.Context, url *model.URL) (*model.URL, error)

	// InsertList adds multiple URL entries to the repository in a single operation.
	// Returns the list of created URL models or an error.
	InsertList(ctx context.Context, urls []*model.URL) ([]*model.URL, error)

	// FindByID retrieves a URL entry by its short identifier.
	// Returns the corresponding URL model or an error if not found.
	FindByID(ctx context.Context, shortID string) (*model.URL, error)

	// FindListByUserID retrieves all URL entries associated with a specific user.
	// Returns a slice of URL models or an error if retrieval fails.
	FindListByUserID(ctx context.Context, userID string) ([]*model.URL, error)

	// DeleteListByUserIDAndShortIDs removes multiple URL entries for a specific user
	// based on their user ID and a list of short identifiers. Returns an error if the operation fails.
	DeleteListByUserIDAndShortIDs(ctx context.Context, userID string, shortIDs []string) error

	// List retrieves all URL entries from the repository.
	// Returns a slice of URL models or an error if retrieval fails.
	List(ctx context.Context) ([]*model.URL, error)

	// Ping checks the repository's connection health.
	// Returns an error if the repository is unreachable.
	Ping(ctx context.Context) error
}
