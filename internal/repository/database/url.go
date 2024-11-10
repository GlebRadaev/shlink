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
	_, err := db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS urls (
            id SERIAL PRIMARY KEY,
            short_id VARCHAR(8) UNIQUE NOT NULL,
            original_url VARCHAR(2048) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		log.Printf("Failed to create table 'urls': %v", err)
	} else {
		log.Print("Table 'urls' created successfully.")
	}
	return &URLRepository{db: db}
}

func (r *URLRepository) Insert(ctx context.Context, model *model.URL) (*model.URL, error) {
	query := `INSERT INTO urls (short_id, original_url) VALUES ($1, $2) RETURNING id, short_id, original_url, created_at`
	err := r.db.QueryRow(ctx, query, model.ShortID, model.OriginalURL).
		Scan(&model.ID, &model.ShortID, &model.OriginalURL, &model.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert URL: %v", err)
	}
	return model, nil
}

func (r *URLRepository) FindByID(ctx context.Context, shortID string) (*model.URL, error) {
	query := `SELECT id, short_id, original_url, created_at FROM urls WHERE short_id = $1`
	url := &model.URL{}
	err := r.db.QueryRow(ctx, query, shortID).Scan(&url.ID, &url.ShortID, &url.OriginalURL, &url.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return url, nil
}

func (r *URLRepository) List(ctx context.Context) ([]*model.URL, error) {
	query := `SELECT id, short_id, original_url, created_at FROM urls`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find URLs: %v", err)
	}
	defer rows.Close()

	var urls []*model.URL
	for rows.Next() {
		url := &model.URL{}
		if err := rows.Scan(&url.ID, &url.ShortID, &url.OriginalURL, &url.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan URL data: %v", err)
		}
		urls = append(urls, url)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to find URLs: error occurred during rows iteration: %v", rows.Err())
	}
	return urls, nil
}
