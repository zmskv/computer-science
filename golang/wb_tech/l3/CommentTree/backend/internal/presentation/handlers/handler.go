package handlers

import (
	"net/http"
	"strconv"

	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/presentation/dto"
	"go.uber.org/zap"
)

type Handler struct {
	service interfaces.CommentService
	logger  *zap.Logger
}

func NewHandler(service interfaces.CommentService, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreateComment(c *ginext.Context) {
	var req dto.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(400, map[string]string{"error": err.Error()})
		return
	}

	if req.Text == "" {
		c.AbortWithStatusJSON(400, map[string]string{"error": "text is required"})
		return
	}

	if req.Author == "" {
		c.AbortWithStatusJSON(400, map[string]string{"error": "author is required"})
		return
	}

	id, err := h.service.CreateComment(c.Request.Context(), req.ParentID, req.Text, req.Author)
	if err != nil {
		h.logger.Error("failed to create comment", zap.Error(err))
		c.AbortWithStatusJSON(500, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(201, map[string]string{"id": id})
}

func (h *Handler) GetComments(c *ginext.Context) {
	params := dto.GetCommentParams{
		ParentID: c.Query("parent"),
		Search:   c.Query("search"),
		SortBy:   c.Query("sort"),
	}

	pageStr := c.Query("page")
	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil && page > 0 {
			params.Page = page - 1
		}
	}

	pageSizeStr := c.Query("page_size")
	if pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err == nil && pageSize > 0 {
			params.PageSize = pageSize
		}
	} else {
		params.PageSize = 50
	}

	comments, err := h.service.GetComment(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("failed to get comments", zap.Error(err))
		c.AbortWithStatusJSON(500, map[string]string{"error": err.Error()})
		return
	}

	c.JSON(200, comments)
}

func (h *Handler) DeleteComment(c *ginext.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(400, map[string]string{"error": "id is required"})
		return
	}

	err := h.service.DeleteComment(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to delete comment", zap.Error(err))
		c.AbortWithStatusJSON(500, map[string]string{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
