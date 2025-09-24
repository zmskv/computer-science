package interfaces

import (
	"context"
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/domain/entity"
)

type EventRepository interface {
	Create(ctx context.Context, e entity.Event) error
	Update(ctx context.Context, e entity.Event) error
	Delete(ctx context.Context, userID int64, id string) error
	FindByDay(ctx context.Context, userID int64, day time.Time) ([]entity.Event, error)
	FindByWeek(ctx context.Context, userID int64, day time.Time) ([]entity.Event, error)
	FindByMonth(ctx context.Context, userID int64, day time.Time) ([]entity.Event, error)
}
