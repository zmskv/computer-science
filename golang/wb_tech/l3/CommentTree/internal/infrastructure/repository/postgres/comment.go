package postgres

import (
	"context"

	"github.com/wb-go/wbf/dbpg"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/entity"
	"go.uber.org/zap"
)

type CommentRepository struct {
	db     *dbpg.DB
	logger *zap.Logger
}

func NewCommentRepository(db *dbpg.DB, logger *zap.Logger) *CommentRepository {
	return &CommentRepository{db: db, logger: logger}
}

func (r *CommentRepository) CreateComment(ctx context.Context, comment entity.Comment) error {
	
}
