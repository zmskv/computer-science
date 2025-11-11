package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/infrastructure/repository/dto"
)

type CommentRepository interface {
	GetComment(ctx context.Context, params dto.GetCommentParams) ([]entity.Comment, error)
	CreateComment(ctx context.Context, comment entity.Comment) error
	EditComment(ctx context.Context, id, text string) error
	DeleteComment(ctx context.Context, id string) error
}
