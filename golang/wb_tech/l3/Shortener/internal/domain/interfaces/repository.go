package interfaces

import (
	"context"
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/domain/entity"
)

type ShortURLRepository interface {
	Save(ctx context.Context, u entity.ShortURL) error
	GetByCode(ctx context.Context, code string) (entity.ShortURL, error)
}

type AnalyticsRepository interface {
	SaveClick(ctx context.Context, click entity.ClickEvent) error
	GetClicks(ctx context.Context, code string) ([]entity.ClickEvent, error)
	CountByDay(ctx context.Context, code string) (map[time.Time]int, error)
	CountByMonth(ctx context.Context, code string) (map[string]int, error)
	CountByUserAgent(ctx context.Context, code string) (map[string]int, error)
}
