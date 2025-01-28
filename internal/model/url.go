package model

import "time"

// URL represents a shortened URL record in the database.
type URL struct {
	ID          int       `db:"id"`           // ID is the primary key for the URL record.
	ShortID     string    `db:"short_id"`     // ShortID is the unique identifier for the shortened URL.
	OriginalURL string    `db:"original_url"` // OriginalURL is the full URL before shortening.
	UserID      string    `db:"user_id"`      // UserID is the identifier for the user who created the shortened URL.
	CreatedAt   time.Time `db:"created_at"`   // CreatedAt is the timestamp when the shortened URL was created.
	DeletedFlag bool      `db:"is_deleted"`   // DeletedFlag indicates if the URL is marked as deleted.
}
