package dto

import "time"

type ShortURLRequest struct {
	OriginalURL string `json:"original_url" validate:"required,url"`
}

type ShortURLResponse struct {
	ID          string    `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	CreatedAt   time.Time `json:"created_at"`
}

type ClickEventResponse struct {
	ID        string    `json:"id"`
	ShortCode string    `json:"short_code"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type AnalyticsResponse struct {
	Total      int                    `json:"total"`
	ByDay      map[string]int         `json:"by_day"`
	ByMonth    map[string]int         `json:"by_month"`
	ByUserAgent map[string]int        `json:"by_user_agent"`
	Events     []ClickEventResponse   `json:"events"`
}
