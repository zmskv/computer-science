package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/interfaces"
	"go.uber.org/zap"
)

type Handler struct {
	service interfaces.NotifierService
	logger  *zap.Logger
}

func NewHandler(service interfaces.NotifierService, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Notify(c *ginext.Context) {
	var req struct {
		Title          string    `json:"title" binding:"required"`
		Body           string    `json:"body" binding:"required"`
		ExpirationTime time.Time `json:"expiration_time" binding:"required"`
		Email          string    `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	note := entity.Note{
		ID:             uuid.NewString(),
		Title:          req.Title,
		Body:           req.Body,
		ExpirationTime: req.ExpirationTime.UTC(),
		Recipient:      req.Email,
		Channel:        "email",
		Status:         "pending",
	}

	id, err := h.service.CreateNote(context.Background(), note)
	if err != nil {
		h.logger.Error("failed to create note", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, ginext.H{"id": id})
}

func (h *Handler) GetNotifyById(c *ginext.Context) {
	id := c.Param("id")
	note, err := h.service.GetNote(context.Background(), id)
	if err != nil {
		h.logger.Error("failed to get note", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, ginext.H{"error": "note not found"})
		return
	}

	c.JSON(http.StatusOK, note)
}

func (h *Handler) DeleteNotifyById(c *ginext.Context) {
	id := c.Param("id")
	err := h.service.DeleteNote(context.Background(), id)
	if err != nil {
		h.logger.Error("failed to delete note", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to delete note"})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"status": "deleted"})
}
