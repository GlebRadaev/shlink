package model

import "time"

type URL struct {
	ID          int       `db:"id"`
	ShortID     string    `db:"short_id"`
	OriginalURL string    `db:"original_url"`
	UserID      string    `db:"user_id"`
	CreatedAt   time.Time `db:"created_at"`
}
