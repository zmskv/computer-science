package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"calendar/internal/domain/entity"
	"calendar/internal/domain/interfaces"
	"calendar/internal/presentation/http/ginapp/dto"

	"github.com/gin-gonic/gin"
)

var acceptedDateLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04",
	"2006-01-02",
}

type Handler struct {
	service interfaces.EventService
	logger  interfaces.Logger
}

func NewHandler(service interfaces.EventService, logger interfaces.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var req dto.CreateEventReq
	if err := c.ShouldBind(&req); err != nil {
		h.logger.Error("create_event bind failed", map[string]any{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	date, err := parseDate(req.Date)
	if err != nil {
		h.logger.Error("create_event invalid date", map[string]any{"date": req.Date})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}
	remindAt, err := parseOptionalDate(req.RemindAt)
	if err != nil {
		h.logger.Error("create_event invalid remind_at", map[string]any{"remind_at": req.RemindAt})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid remind_at format"})
		return
	}

	e, err := h.service.CreateEvent(c.Request.Context(), req.UserID, date, req.Title, remindAt)
	if err != nil {
		h.logger.Error("create_event failed", map[string]any{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"result": e})
}

func (h *Handler) UpdateEvent(c *gin.Context) {
	var req dto.UpdateEventReq
	if err := c.ShouldBind(&req); err != nil {
		h.logger.Error("update_event bind failed", map[string]any{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	date, err := parseDate(req.Date)
	if err != nil {
		h.logger.Error("update_event invalid date", map[string]any{"date": req.Date})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}
	remindAt, err := parseOptionalDate(req.RemindAt)
	if err != nil {
		h.logger.Error("update_event invalid remind_at", map[string]any{"remind_at": req.RemindAt})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid remind_at format"})
		return
	}

	e, err := h.service.UpdateEvent(c.Request.Context(), req.ID, req.UserID, date, req.Title, remindAt)
	if err != nil {
		h.logger.Error("update_event failed", map[string]any{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": e})
}

func (h *Handler) DeleteEvent(c *gin.Context) {
	var req dto.DeleteEventReq
	if err := c.ShouldBind(&req); err != nil {
		h.logger.Error("delete_event bind failed", map[string]any{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	if err := h.service.DeleteEvent(c.Request.Context(), req.UserID, req.ID); err != nil {
		h.logger.Error("delete_event failed", map[string]any{"error": err.Error()})
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "deleted"})
}

func (h *Handler) EventsForDay(c *gin.Context) {
	h.queryEvents(c, "day")
}

func (h *Handler) EventsForWeek(c *gin.Context) {
	h.queryEvents(c, "week")
}

func (h *Handler) EventsForMonth(c *gin.Context) {
	h.queryEvents(c, "month")
}

func (h *Handler) ArchivedEvents(c *gin.Context) {
	var req dto.ArchivedQueryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("archived_events bind failed", map[string]any{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	events, err := h.service.ArchivedEvents(c.Request.Context(), req.UserID)
	if err != nil {
		h.logger.Error("archived_events failed", map[string]any{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": events})
}

func (h *Handler) queryEvents(c *gin.Context, kind string) {
	var req dto.QueryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("query_events bind failed", map[string]any{"error": err.Error(), "kind": kind})
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id or date"})
		return
	}
	date, err := parseDate(req.Date)
	if err != nil {
		h.logger.Error("query_events invalid date", map[string]any{"kind": kind, "date": req.Date})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	var events []entity.Event
	switch kind {
	case "day":
		events, err = h.service.EventsForDay(c.Request.Context(), req.UserID, date)
	case "week":
		events, err = h.service.EventsForWeek(c.Request.Context(), req.UserID, date)
	case "month":
		events, err = h.service.EventsForMonth(c.Request.Context(), req.UserID, date)
	}

	if err != nil {
		h.logger.Error("query_events failed", map[string]any{"kind": kind, "error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": events})
}

func parseDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	for _, layout := range acceptedDateLayouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format: %s", raw)
}

func parseOptionalDate(raw string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	parsed, err := parseDate(raw)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
