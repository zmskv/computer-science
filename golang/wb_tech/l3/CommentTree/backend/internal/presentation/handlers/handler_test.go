package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/interfaces/mocks"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/presentation/dto"
	"go.uber.org/zap"
)

func setupRouter(handler *Handler) *ginext.Engine {
	router := ginext.New("")

	router.POST("/comments", handler.CreateComment)
	router.GET("/comments", handler.GetComments)
	router.DELETE("/comments/:id", handler.DeleteComment)

	return router
}

func TestHandler_CreateComment(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mocks.MockCommentService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful creation",
			requestBody: dto.CommentRequest{
				Text:   "Test comment",
				Author: "Test Author",
			},
			setupMock: func(m *mocks.MockCommentService) {
				m.EXPECT().CreateComment(gomock.Any(), "", "Test comment", "Test Author").Return("test-id", nil)
			},
			expectedStatus: 201,
			expectedBody: map[string]interface{}{
				"id": "test-id",
			},
		},
		{
			name: "successful creation with parent",
			requestBody: dto.CommentRequest{
				ParentID: "parent-id",
				Text:     "Reply comment",
				Author:   "Reply Author",
			},
			setupMock: func(m *mocks.MockCommentService) {
				m.EXPECT().CreateComment(gomock.Any(), "parent-id", "Reply comment", "Reply Author").Return("reply-id", nil)
			},
			expectedStatus: 201,
			expectedBody: map[string]interface{}{
				"id": "reply-id",
			},
		},
		{
			name: "missing text",
			requestBody: dto.CommentRequest{
				Author: "Test Author",
			},
			setupMock:      func(m *mocks.MockCommentService) {},
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "text is required",
			},
		},
		{
			name: "missing author",
			requestBody: dto.CommentRequest{
				Text: "Test comment",
			},
			setupMock:      func(m *mocks.MockCommentService) {},
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "author is required",
			},
		},
		{
			name: "service error",
			requestBody: dto.CommentRequest{
				Text:   "Test comment",
				Author: "Test Author",
			},
			setupMock: func(m *mocks.MockCommentService) {
				m.EXPECT().CreateComment(gomock.Any(), "", "Test comment", "Test Author").Return("", errors.New("service error"))
			},
			expectedStatus: 500,
			expectedBody: map[string]interface{}{
				"error": "service error",
			},
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			setupMock:      func(m *mocks.MockCommentService) {},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockCommentService(ctrl)
			tt.setupMock(mockService)

			logger := zap.NewNop()
			handler := NewHandler(mockService, logger)
			router := setupRouter(handler)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/comments", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestHandler_GetComments(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*mocks.MockCommentService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "successful get all",
			queryParams: "",
			setupMock: func(m *mocks.MockCommentService) {
				comments := []entity.Comment{
					{Id: "1", Text: "Comment 1", Author: "Author 1", Date: time.Now()},
					{Id: "2", Text: "Comment 2", Author: "Author 2", Date: time.Now()},
				}
				m.EXPECT().GetComment(gomock.Any(), gomock.Any()).Return(comments, nil)
			},
			expectedStatus: 200,
			expectedCount:  2,
		},
		{
			name:        "successful get with parent",
			queryParams: "?parent=parent-id",
			setupMock: func(m *mocks.MockCommentService) {
				comments := []entity.Comment{
					{Id: "1", ParentID: "parent-id", Text: "Reply 1", Author: "Author 1", Date: time.Now()},
				}
				m.EXPECT().GetComment(gomock.Any(), gomock.Any()).Return(comments, nil)
			},
			expectedStatus: 200,
			expectedCount:  1,
		},
		{
			name:        "successful get with pagination",
			queryParams: "?page=2&page_size=10",
			setupMock: func(m *mocks.MockCommentService) {
				comments := []entity.Comment{}
				m.EXPECT().GetComment(gomock.Any(), gomock.Any()).Return(comments, nil)
			},
			expectedStatus: 200,
			expectedCount:  0,
		},
		{
			name:        "service error",
			queryParams: "",
			setupMock: func(m *mocks.MockCommentService) {
				m.EXPECT().GetComment(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockCommentService(ctrl)
			tt.setupMock(mockService)

			logger := zap.NewNop()
			handler := NewHandler(mockService, logger)
			router := setupRouter(handler)

			req := httptest.NewRequest("GET", "/comments"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCount > 0 {
				var response []entity.Comment
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, tt.expectedCount)
			}
		})
	}
}

func TestHandler_DeleteComment(t *testing.T) {
	tests := []struct {
		name           string
		commentID      string
		setupMock      func(*mocks.MockCommentService)
		expectedStatus int
	}{
		{
			name:      "successful deletion",
			commentID: "comment-id",
			setupMock: func(m *mocks.MockCommentService) {
				m.EXPECT().DeleteComment(gomock.Any(), "comment-id").Return(nil)
			},
			expectedStatus: 204,
		},
		{
			name:           "missing id",
			commentID:      "",
			setupMock:      func(m *mocks.MockCommentService) {},
			expectedStatus: 404,
		},
		{
			name:      "service error",
			commentID: "comment-id",
			setupMock: func(m *mocks.MockCommentService) {
				m.EXPECT().DeleteComment(gomock.Any(), "comment-id").Return(errors.New("service error"))
			},
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockCommentService(ctrl)
			tt.setupMock(mockService)

			logger := zap.NewNop()
			handler := NewHandler(mockService, logger)
			router := setupRouter(handler)

			url := "/comments/" + tt.commentID
			if tt.commentID == "" {
				url = "/comments/"
			}
			req := httptest.NewRequest("DELETE", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
