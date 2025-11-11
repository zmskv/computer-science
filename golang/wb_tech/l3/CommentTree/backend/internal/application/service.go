package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/interfaces"
	repositoryDTO "github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/infrastructure/repository/dto"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/presentation/dto"
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

func (s *CommentService) GetComment(ctx context.Context, params dto.GetCommentParams) ([]entity.Comment, error) {
	repositoryParams := repositoryDTO.GetCommentParams{
		ParentID: params.ParentID,
		Search:   params.Search,
		SortBy:   params.SortBy,
		Page:     params.Page,
		PageSize: params.PageSize,
	}
	comments, err := s.repo.GetComment(ctx, repositoryParams)
	if err != nil {
		s.logger.Error("failed to get comments", zap.Error(err))
		return nil, err
	}
	return comments, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, id string) error {
	err := s.repo.DeleteComment(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete comment", zap.Error(err))
		return err
	}
	s.logger.Info("Comment successfully deleted", zap.String("id", id))
	return nil
}

func (s *CommentService) EditComment(ctx context.Context, id, text string) error {
	err := s.repo.EditComment(ctx, id, text)
	if err != nil {
		s.logger.Error("failed to edit comment", zap.Error(err))
		return err
	}
	s.logger.Info("Comment successfully edited", zap.String("id", id))
	return nil
}
