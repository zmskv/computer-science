package interfaces

import (
	"context"
	"time"

	"calendar/internal/domain/entity"
)

type EventService interface {
	CreateEvent(ctx context.Context, userID int64, date time.Time, title string, reminderAt *time.Time) (entity.Event, error)
	UpdateEvent(ctx context.Context, id string, userID int64, date time.Time, title string, reminderAt *time.Time) (entity.Event, error)
	DeleteEvent(ctx context.Context, userID int64, id string) error
	EventsForWeek(ctx context.Context, userID int64, date time.Time) ([]entity.Event, error)
	EventsForDay(ctx context.Context, userID int64, date time.Time) ([]entity.Event, error)
	EventsForMonth(ctx context.Context, userID int64, date time.Time) ([]entity.Event, error)
	ArchivedEvents(ctx context.Context, userID int64) ([]entity.Event, error)
}
