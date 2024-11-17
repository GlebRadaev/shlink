package database

import (
	"context"
	"fmt"
	"log"

	"github.com/GlebRadaev/shlink/internal/interfaces"
	"github.com/GlebRadaev/shlink/internal/model"
	"github.com/jackc/pgx/v5"
)

type URLRepository struct {
	db interfaces.DBPool
}

func NewURLRepository(db interfaces.DBPool) interfaces.IURLRepository {
	return &URLRepository{db: db}
}

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

func (r *URLRepository) FindByID(ctx context.Context, shortID string) (*model.URL, error) {
	query := `
		SELECT id, short_id, original_url, user_id, created_at FROM urls 
		WHERE short_id = $1`
	url := &model.URL{}
	err := r.db.QueryRow(ctx, query, shortID).Scan(&url.ID, &url.ShortID, &url.OriginalURL, &url.UserID, &url.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return url, nil
}

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

func (r *URLRepository) Ping(ctx context.Context) error {
	query := `SELECT 1`
	var result int
	err := r.db.QueryRow(ctx, query).Scan(&result)
	if err != nil {
		return err
	}
	return nil
}
