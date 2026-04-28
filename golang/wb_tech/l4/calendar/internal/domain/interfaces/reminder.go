package interfaces

import (
	"context"

	"calendar/internal/domain/entity"
)

type ReminderSender interface {
	Send(ctx context.Context, event entity.Event) error
}
