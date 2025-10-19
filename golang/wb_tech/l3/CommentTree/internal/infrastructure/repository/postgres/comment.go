package postgres

import (
	"context"

	"github.com/wb-go/wbf/dbpg"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/infrastructure/repository/dto"
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
	var query string
	commentDTO := dto.Comment{
		Id:       comment.Id,
		ParentID: comment.ParentID,
		Text:     comment.Text,
		Author:   comment.Author,
		Date:     comment.Date,
	}
	query = `INSERT INTO comments (id, parent_id, text, author, date) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, commentDTO.Id, commentDTO.ParentID, commentDTO.Text, commentDTO.Author, commentDTO.Date)
	if err != nil {
		r.logger.Error("failed to create comment", zap.Error(err))
		return err
	}
	return nil

}

func (r *CommentRepository) EditComment(ctx context.Context, id, text string) error {
	var query string
	query = `UPDATE comments SET text = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, text, id)
	if err != nil {
		r.logger.Error("failed to edit comment", zap.Error(err))
		return err
	}
	return nil

}

func (r *CommentRepository) GetComment(ctx context.Context, id string) ([]entity.Comment, error) {

}

func (r *CommentRepository) DeleteComment(ctx context.Context, id string) error {

}
