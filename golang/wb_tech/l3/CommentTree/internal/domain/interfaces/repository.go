package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/entity"
)

type CommentRepository interface {
	GetComments(ctx context.Context, parent string) ([]entity.Comment, error)
	GetComment(ctx context.Context, id string) (entity.Comment, error)
	CreateComment(ctx context.Context, comment entity.Comment) error
	EditComment(ctx context.Context, comment entity.Comment) error
	DeleteComment(ctx context.Context, id string) error
}
