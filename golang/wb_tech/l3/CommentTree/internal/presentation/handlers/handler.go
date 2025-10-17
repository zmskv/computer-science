package handlers

import (
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
		c.AbortWithStatusJSON(400, err.Error())
		return
	}

	comment, err := h.service.CreateComment(c.Request.Context(), req.ParentID, req.Text, req.Author)
	if err != nil {
		c.AbortWithStatusJSON(500, err.Error())
		return
	}
	c.JSON(201, comment)

}

func (h *Handler) GetComments(c *ginext.Context) {}

func (h *Handler) DeleteComment(c *ginext.Context) {}
