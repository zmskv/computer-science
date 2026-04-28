package interfaces

import (
	"context"
	"time"

	"calendar/internal/domain/entity"
)

type EventRepository interface {
	Create(ctx context.Context, e entity.Event) error
	Update(ctx context.Context, e entity.Event) error
	Delete(ctx context.Context, userID int64, id string) error
	GetByID(ctx context.Context, id string) (entity.Event, error)
	FindByDay(ctx context.Context, userID int64, day time.Time) ([]entity.Event, error)
	FindByWeek(ctx context.Context, userID int64, day time.Time) ([]entity.Event, error)
	FindByMonth(ctx context.Context, userID int64, day time.Time) ([]entity.Event, error)
	ListArchived(ctx context.Context, userID int64) ([]entity.Event, error)
	ArchiveBefore(ctx context.Context, before time.Time) ([]entity.Event, error)
	MarkReminderSent(ctx context.Context, id string) error
}
