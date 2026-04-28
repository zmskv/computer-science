package reminder

import (
	"context"
	"time"

	"calendar/internal/domain/entity"
	"calendar/internal/domain/interfaces"
)

type LogNotifier struct {
	logger interfaces.Logger
}

func NewLogNotifier(logger interfaces.Logger) *LogNotifier {
	return &LogNotifier{logger: logger}
}

func (n *LogNotifier) Send(ctx context.Context, event entity.Event) error {
	n.logger.Info("sending reminder", map[string]any{
		"event_id":  event.ID,
		"user_id":   event.UserID,
		"title":     event.Title,
		"event_at":  event.Date.Format(timeLayout),
		"remind_at": formatOptionalTime(event.ReminderAt),
	})
	return nil
}

const timeLayout = time.RFC3339

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(timeLayout)
}
