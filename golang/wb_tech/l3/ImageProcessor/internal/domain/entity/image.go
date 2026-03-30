package entity

import "time"

type ImageStatus string

const (
	StatusQueued     ImageStatus = "queued"
	StatusProcessing ImageStatus = "processing"
	StatusReady      ImageStatus = "ready"
	StatusFailed     ImageStatus = "failed"
)

type Image struct {
	ID               string      `json:"id"`
	OriginalFilename string      `json:"original_filename"`
	Format           string      `json:"format"`
	Status           ImageStatus `json:"status"`
	OriginalPath     string      `json:"original_path,omitempty"`
	ProcessedPath    string      `json:"processed_path,omitempty"`
	ThumbnailPath    string      `json:"thumbnail_path,omitempty"`
	Error            string      `json:"error,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

type ProcessingOptions struct {
	MaxWidth      int
	MaxHeight     int
	ThumbnailSize int
	WatermarkText string
}

type ProcessedImage struct {
	Processed []byte
	Thumbnail []byte
	Format    string
}
