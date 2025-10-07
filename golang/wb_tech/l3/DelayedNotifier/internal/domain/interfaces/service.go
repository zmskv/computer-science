package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/entity"
)

type NotifierService interface {
	CreateNote(ctx context.Context, note entity.Note) (string, error)
	GetNote(ctx context.Context, id string) (entity.Note, error)
	DeleteNote(ctx context.Context, id string) error
	SendNotification(ctx context.Context, note entity.Note) error
	UpdateNoteStatus(ctx context.Context, id, status string) error
	UpdateNoteRetries(ctx context.Context, id string, retries int) error
}
