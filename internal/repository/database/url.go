// Package database implements the database logic for managing URL data.
// It provides functionality for inserting, querying, updating, and deleting URL records in a PostgreSQL database.
// The repository interacts with the database through the DBPool interface, typically using pgx package for PostgreSQL operations.
package database

import (
	"context"
	"fmt"
	"log"

	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
)

// URLRepository represents a repository for URL data in the database.
type URLRepository struct {
	db interfaces.DBPool
}

// NewURLRepository creates a new instance of URLRepository with the provided DBPool.
func NewURLRepository(db interfaces.DBPool) interfaces.IURLRepository {
	return &URLRepository{db: db}
}

// Insert inserts a new URL into the database, or updates the existing one based on the original URL.
// If the URL already exists, it updates the short ID. Returns the inserted or updated URL.
func (r *URLRepository) Insert(ctx context.Context, url *model.URL) (*model.URL, error) {
	query := `
		INSERT INTO urls (short_id, original_url, user_id) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (original_url) DO UPDATE 
		SET short_id = urls.short_id 
		RETURNING id, short_id, original_url, user_id, created_at`
	err := r.db.QueryRow(ctx, query, url.ShortID, url.OriginalURL, url.UserID).
		Scan(&url.ID, &url.ShortID, &url.OriginalURL, &url.UserID, &url.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert URL: %v", err)
	}
	return url, nil
}

// InsertList inserts a list of URLs into the database. It uses a transaction to insert URLs one by one and commits the transaction.
func (r *URLRepository) InsertList(ctx context.Context, urls []*model.URL) ([]*model.URL, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}

	var result []*model.URL
	for _, url := range urls {
		insertedURL, err := r.Insert(ctx, url)
		if err != nil {
			return nil, err
		}
		result = append(result, insertedURL)
	}
	_ = tx.Commit(ctx)
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Printf("Failed to rollback transaction: %v", err)
		}
	}()
	return result, nil
}

// FindByID finds a URL by its short ID. Returns the URL if found, otherwise returns nil.
func (r *URLRepository) FindByID(ctx context.Context, shortID string) (*model.URL, error) {
	query := `
		SELECT id, short_id, original_url, user_id, created_at, is_deleted FROM urls 
		WHERE short_id = $1`
	url := &model.URL{}
	err := r.db.QueryRow(ctx, query, shortID).Scan(&url.ID, &url.ShortID, &url.OriginalURL, &url.UserID, &url.CreatedAt, &url.DeletedFlag)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return url, nil
}

// FindListByUserID finds all URLs associated with a specific user. It returns a list of URLs.
func (r *URLRepository) FindListByUserID(ctx context.Context, userID string) ([]*model.URL, error) {
	query := `
		SELECT id, short_id, original_url, user_id, created_at FROM urls 
		WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*model.URL
	for rows.Next() {
		url := &model.URL{}
		err := rows.Scan(&url.ID, &url.ShortID, &url.OriginalURL, &url.UserID, &url.CreatedAt)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

// List retrieves all URLs in the database. It returns a list of all URLs stored.
func (r *URLRepository) List(ctx context.Context) ([]*model.URL, error) {
	query := `
		SELECT id, short_id, original_url, created_at, user_id 
		FROM urls`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find URLs: %v", err)
	}
	defer rows.Close()

	var urls []*model.URL
	for rows.Next() {
		url := &model.URL{}
		if err := rows.Scan(&url.ID, &url.ShortID, &url.OriginalURL, &url.CreatedAt, &url.UserID); err != nil {
			return nil, fmt.Errorf("failed to scan URL data: %v", err)
		}
		urls = append(urls, url)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to find URLs: error occurred during rows iteration: %v", rows.Err())
	}
	return urls, nil
}

// Ping checks if the database is reachable by executing a simple query.
func (r *URLRepository) Ping(ctx context.Context) error {
	query := `SELECT 1`
	var result int
	err := r.db.QueryRow(ctx, query).Scan(&result)
	if err != nil {
		return err
	}
	return nil
}

// DeleteListByUserIDAndShortIDs soft deletes URLs by marking them as deleted based on userID and shortID list.
// It performs the deletion inside a transaction to ensure consistency.
func (r *URLRepository) DeleteListByUserIDAndShortIDs(ctx context.Context, userID string, shortIDs []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	query := `
		UPDATE urls
		SET is_deleted = true
		WHERE user_id = $1 AND short_id = ANY($2)
	`
	_, err = tx.Exec(ctx, query, userID, pq.Array(shortIDs))
	if err != nil {
		_ = tx.Rollback(ctx) // Игнорируем ошибку, но явным образом
		return fmt.Errorf("failed to delete short urls for user: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		_ = tx.Rollback(ctx) // Игнорируем ошибку, но явным образом
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	log.Printf("Successfully marked URLs as deleted for userID=%s: %v", userID, shortIDs)
	return nil
}
