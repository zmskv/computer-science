package entity

import "time"

type ShortURL struct {
	ID          string
	OriginalURL string
	ShortCode   string
	CreatedAt   time.Time
}

type ClickEvent struct {
	ID        string
	ShortCode string
	UserAgent string
	CreatedAt time.Time
}
