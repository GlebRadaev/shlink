package interfaces

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBPool defines the interface for a database connection pool, providing methods for querying, executing, and transaction handling.
type DBPool interface {
	// QueryRow executes a query that is expected to return at most one row.
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row

	// Query executes a query that is expected to return multiple rows.
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)

	// Exec executes a query without returning any rows, typically for INSERT, UPDATE, or DELETE operations.
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)

	// Begin starts a new transaction.
	Begin(ctx context.Context) (pgx.Tx, error)

	// Close closes the database connection pool.
	Close()
}
