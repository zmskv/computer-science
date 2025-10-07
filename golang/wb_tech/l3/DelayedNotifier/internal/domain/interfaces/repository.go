package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/entity"
)

type NotifierRepository interface {
	CreateNote(ctx context.Context, note entity.Note) (string, error)
	GetNote(ctx context.Context, id string) (entity.Note, error)
	DeleteNote(ctx context.Context, id string) error
	UpdateNoteStatus(ctx context.Context, id, status string) error
	UpdateNoteRetries(ctx context.Context, id string, retries int) error
	RemoveFromSchedule(ctx context.Context, ids ...string) error
	GetDueNotificationIDs(ctx context.Context) ([]string, error)
}
