package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/entity"
)

type CommentRepository interface {
	GetComment(ctx context.Context, id string) ([]entity.Comment, error)
	CreateComment(ctx context.Context, comment entity.Comment) error
	EditComment(ctx context.Context, text string) error
	DeleteComment(ctx context.Context, id string) error
}
