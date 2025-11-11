package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	commentDTO := dto.Comment{
		Id:       comment.Id,
		ParentID: comment.ParentID,
		Text:     comment.Text,
		Author:   comment.Author,
		Date:     comment.Date,
	}

	var parentID interface{}
	if commentDTO.ParentID == "" {
		parentID = nil
	} else {
		parentID = commentDTO.ParentID
	}

	query := `INSERT INTO comments (id, parent_id, text, author, date) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, commentDTO.Id, parentID, commentDTO.Text, commentDTO.Author, commentDTO.Date)
	if err != nil {
		r.logger.Error("failed to create comment", zap.Error(err))
		return err
	}
	return nil
}

func (r *CommentRepository) EditComment(ctx context.Context, id, text string) error {
	query := `UPDATE comments SET text = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, text, id)
	if err != nil {
		r.logger.Error("failed to edit comment", zap.Error(err))
		return err
	}
	return nil

}

func (r *CommentRepository) GetComment(ctx context.Context, params dto.GetCommentParams) ([]entity.Comment, error) {
	var query strings.Builder
	var args []interface{}
	argPos := 1

	if params.Search != "" {
		query.WriteString(`
			WITH RECURSIVE comment_tree AS (
				SELECT id, parent_id, text, author, date, 0 as level
				FROM comments
				WHERE text ILIKE $` + fmt.Sprintf("%d", argPos))
		args = append(args, "%"+params.Search+"%")
		argPos++
		query.WriteString(`
				UNION ALL
				SELECT c.id, c.parent_id, c.text, c.author, c.date, ct.level + 1
				FROM comments c
				INNER JOIN comment_tree ct ON c.parent_id = ct.id
			)
			SELECT id, parent_id, text, author, date FROM comment_tree`)
		if params.ParentID != "" {
			query.WriteString(fmt.Sprintf(" WHERE id = $%d OR parent_id = $%d", argPos, argPos))
			args = append(args, params.ParentID)
		}
	} else {
		query.WriteString(`
			WITH RECURSIVE comment_tree AS (
				SELECT id, parent_id, text, author, date, 0 as level
				FROM comments
				WHERE `)
		if params.ParentID != "" {
			query.WriteString(fmt.Sprintf("id = $%d", argPos))
			args = append(args, params.ParentID)
			argPos++
		} else {
			query.WriteString("parent_id IS NULL")
		}
		query.WriteString(`
				UNION ALL
				SELECT c.id, c.parent_id, c.text, c.author, c.date, ct.level + 1
				FROM comments c
				INNER JOIN comment_tree ct ON c.parent_id = ct.id
			)
			SELECT id, parent_id, text, author, date FROM comment_tree`)
	}

	if params.SortBy != "" {
		switch params.SortBy {
		case "date_asc":
			query.WriteString(" ORDER BY date ASC")
		case "date_desc":
			query.WriteString(" ORDER BY date DESC")
		case "author":
			query.WriteString(" ORDER BY author ASC")
		default:
			query.WriteString(" ORDER BY date DESC")
		}
	} else {
		query.WriteString(" ORDER BY date DESC")
	}

	if params.PageSize > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT $%d", argPos))
		args = append(args, params.PageSize)
		argPos++
		if params.Page > 0 {
			query.WriteString(fmt.Sprintf(" OFFSET $%d", argPos))
			args = append(args, params.Page*params.PageSize)
		}
	}

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		r.logger.Error("failed to get comments", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var comments []entity.Comment
	for rows.Next() {
		var commentDTO dto.Comment
		var parentID sql.NullString
		err := rows.Scan(&commentDTO.Id, &parentID, &commentDTO.Text, &commentDTO.Author, &commentDTO.Date)
		if err != nil {
			r.logger.Error("failed to scan comment", zap.Error(err))
			return nil, err
		}
		if parentID.Valid {
			commentDTO.ParentID = parentID.String
		}
		comments = append(comments, entity.Comment{
			Id:       commentDTO.Id,
			ParentID: commentDTO.ParentID,
			Text:     commentDTO.Text,
			Author:   commentDTO.Author,
			Date:     commentDTO.Date,
		})
	}

	return comments, nil
}

func (r *CommentRepository) DeleteComment(ctx context.Context, id string) error {
	query := `
		WITH RECURSIVE comment_tree AS (
			SELECT id FROM comments WHERE id = $1
			UNION ALL
			SELECT c.id FROM comments c
			INNER JOIN comment_tree ct ON c.parent_id = ct.id
		)
		DELETE FROM comments WHERE id IN (SELECT id FROM comment_tree)`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete comment", zap.Error(err))
		return err
	}
	return nil
}
