package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/interfaces"
	"go.uber.org/zap"
)

type CommentService struct {
	repo   interfaces.CommentRepository
	logger *zap.Logger
}

func NewCommentService(repo interfaces.CommentRepository, logger *zap.Logger) *CommentService {
	return &CommentService{repo: repo, logger: logger}
}

func (s *CommentService) CreateComment(ctx context.Context, parent_id, text, author string) (string, error) {
	if parent_id == "" {
		parent_id = uuid.NewString()
	}
	id := uuid.NewString()
	comment := entity.Comment{
		Id:       id,
		ParentID: parent_id,
		Text:     text,
		Author:   author,
		Date:     time.Now().UTC(),
	}

	err := s.repo.CreateComment(ctx, comment)
	if err != nil {
		s.logger.Error("failed to create comment", zap.Error(err))
		return "", err
	}
	s.logger.Info("Comment successfully created", zap.String("id", id))
	return id, nil
}
