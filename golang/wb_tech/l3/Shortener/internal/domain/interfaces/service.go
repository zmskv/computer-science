package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/domain/entity"
)

type ShortenerService interface {
	Create(ctx context.Context, originalURL string) (entity.ShortURL, error)
	Resolve(ctx context.Context, shortCode string, userAgent string) (string, error)
	Analytics(ctx context.Context, shortCode string) (map[string]any, error)
}
