package dto

import (
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/entity"
)

type ImageResponse struct {
	ID               string             `json:"id"`
	OriginalFilename string             `json:"original_filename"`
	Format           string             `json:"format"`
	Status           entity.ImageStatus `json:"status"`
	Error            string             `json:"error,omitempty"`
	OriginalURL      string             `json:"original_url,omitempty"`
	ProcessedURL     string             `json:"processed_url,omitempty"`
	ThumbnailURL     string             `json:"thumbnail_url,omitempty"`
	DownloadURL      string             `json:"download_url,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

func FromEntity(imageMeta entity.Image) ImageResponse {
	response := ImageResponse{
		ID:               imageMeta.ID,
		OriginalFilename: imageMeta.OriginalFilename,
		Format:           imageMeta.Format,
		Status:           imageMeta.Status,
		Error:            imageMeta.Error,
		CreatedAt:        imageMeta.CreatedAt,
		UpdatedAt:        imageMeta.UpdatedAt,
	}

	if imageMeta.OriginalPath != "" {
		response.OriginalURL = "/files/" + imageMeta.OriginalPath
	}

	if imageMeta.ProcessedPath != "" {
		response.ProcessedURL = "/files/" + imageMeta.ProcessedPath
		response.DownloadURL = "/image/" + imageMeta.ID
	}

	if imageMeta.ThumbnailPath != "" {
		response.ThumbnailURL = "/files/" + imageMeta.ThumbnailPath
	}

	return response
}
