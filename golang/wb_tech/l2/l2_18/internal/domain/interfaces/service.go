package interfaces

import (
	"context"
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/domain/entity"
)

type EventService interface {
	CreateEvent(ctx context.Context, userID int64, date time.Time, title string) (entity.Event, error)
	UpdateEvent(ctx context.Context, id string, userID int64, date time.Time, title string) (entity.Event, error)
	DeleteEvent(ctx context.Context, userID int64, id string) error
	EventsForWeek(ctx context.Context, userID int64, date time.Time) ([]entity.Event, error)
	EventsForDay(ctx context.Context, userID int64, date time.Time) ([]entity.Event, error)
	EventsForMonth(ctx context.Context, userID int64, date time.Time) ([]entity.Event, error)
}
