package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/presentation/dto"
)

type CommentService interface {
	GetComment(ctx context.Context, params dto.GetCommentParams) ([]entity.Comment, error)
	CreateComment(ctx context.Context, parent_id, text, author string) (string, error)
	EditComment(ctx context.Context, id, text string) error
	DeleteComment(ctx context.Context, id string) error
}
