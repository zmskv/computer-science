package dto

import "time"

type ShortURLDTO struct {
	ID          string    `db:"id"`
	OriginalURL string    `db:"original_url"`
	ShortCode   string    `db:"short_code"`
	CreatedAt   time.Time `db:"created_at"`
}

type ClickEventDTO struct {
	ID        string    `db:"id"`
	ShortCode string    `db:"short_code"`
	UserAgent string    `db:"user_agent"`
	CreatedAt time.Time `db:"created_at"`
}
