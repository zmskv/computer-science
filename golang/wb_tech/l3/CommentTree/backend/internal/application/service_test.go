package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/interfaces/mocks"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/presentation/dto"
	"go.uber.org/zap"
)

func TestCommentService_CreateComment(t *testing.T) {
	tests := []struct {
		name      string
		parentID  string
		text      string
		author    string
		setupMock func(*mocks.MockCommentRepository)
		wantErr   bool
	}{
		{
			name:     "successful creation",
			parentID: "",
			text:     "Test comment",
			author:   "Test Author",
			setupMock: func(m *mocks.MockCommentRepository) {
				m.EXPECT().CreateComment(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, comment entity.Comment) error {
					if comment.Text == "Test comment" && comment.Author == "Test Author" && comment.ParentID == "" {
						return nil
					}
					return errors.New("unexpected comment")
				})
			},
			wantErr: false,
		},
		{
			name:     "successful creation with parent",
			parentID: "parent-id",
			text:     "Reply comment",
			author:   "Reply Author",
			setupMock: func(m *mocks.MockCommentRepository) {
				m.EXPECT().CreateComment(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, comment entity.Comment) error {
					if comment.Text == "Reply comment" && comment.Author == "Reply Author" && comment.ParentID == "parent-id" {
						return nil
					}
					return errors.New("unexpected comment")
				})
			},
			wantErr: false,
		},
		{
			name:     "repository error",
			parentID: "",
			text:     "Test comment",
			author:   "Test Author",
			setupMock: func(m *mocks.MockCommentRepository) {
				m.EXPECT().CreateComment(gomock.Any(), gomock.Any()).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockCommentRepository(ctrl)
			tt.setupMock(mockRepo)

			logger := zap.NewNop()
			service := NewCommentService(mockRepo, logger)

			id, err := service.CreateComment(context.Background(), tt.parentID, tt.text, tt.author)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, id)
			}
		})
	}
}

func TestCommentService_GetComment(t *testing.T) {
	tests := []struct {
		name      string
		params    dto.GetCommentParams
		setupMock func(*mocks.MockCommentRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name: "successful get all comments",
			params: dto.GetCommentParams{
				Page:     0,
				PageSize: 50,
			},
			setupMock: func(m *mocks.MockCommentRepository) {
				comments := []entity.Comment{
					{Id: "1", Text: "Comment 1", Author: "Author 1", Date: time.Now()},
					{Id: "2", Text: "Comment 2", Author: "Author 2", Date: time.Now()},
				}
				m.EXPECT().GetComment(gomock.Any(), gomock.Any()).Return(comments, nil)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "successful get with parent",
			params: dto.GetCommentParams{
				ParentID: "parent-id",
				Page:     0,
				PageSize: 50,
			},
			setupMock: func(m *mocks.MockCommentRepository) {
				comments := []entity.Comment{
					{Id: "1", ParentID: "parent-id", Text: "Reply 1", Author: "Author 1", Date: time.Now()},
				}
				m.EXPECT().GetComment(gomock.Any(), gomock.Any()).Return(comments, nil)
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "repository error",
			params: dto.GetCommentParams{
				Page:     0,
				PageSize: 50,
			},
			setupMock: func(m *mocks.MockCommentRepository) {
				m.EXPECT().GetComment(gomock.Any(), gomock.Any()).Return(nil, errors.New("database error"))
			},
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockCommentRepository(ctrl)
			tt.setupMock(mockRepo)

			logger := zap.NewNop()
			service := NewCommentService(mockRepo, logger)

			comments, err := service.GetComment(context.Background(), tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, comments)
			} else {
				assert.NoError(t, err)
				assert.Len(t, comments, tt.wantCount)
			}
		})
	}
}

func TestCommentService_DeleteComment(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockCommentRepository)
		wantErr   bool
	}{
		{
			name: "successful deletion",
			id:   "comment-id",
			setupMock: func(m *mocks.MockCommentRepository) {
				m.EXPECT().DeleteComment(gomock.Any(), "comment-id").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			id:   "comment-id",
			setupMock: func(m *mocks.MockCommentRepository) {
				m.EXPECT().DeleteComment(gomock.Any(), "comment-id").Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockCommentRepository(ctrl)
			tt.setupMock(mockRepo)

			logger := zap.NewNop()
			service := NewCommentService(mockRepo, logger)

			err := service.DeleteComment(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommentService_EditComment(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		text      string
		setupMock func(*mocks.MockCommentRepository)
		wantErr   bool
	}{
		{
			name: "successful edit",
			id:   "comment-id",
			text: "Updated text",
			setupMock: func(m *mocks.MockCommentRepository) {
				m.EXPECT().EditComment(gomock.Any(), "comment-id", "Updated text").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			id:   "comment-id",
			text: "Updated text",
			setupMock: func(m *mocks.MockCommentRepository) {
				m.EXPECT().EditComment(gomock.Any(), "comment-id", "Updated text").Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockCommentRepository(ctrl)
			tt.setupMock(mockRepo)

			logger := zap.NewNop()
			service := NewCommentService(mockRepo, logger)

			err := service.EditComment(context.Background(), tt.id, tt.text)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
