package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/presentation/http/ginapp/dto"

	"go.uber.org/zap"
)

const dateLayout = "2006-01-02"

type Handler struct {
	service interfaces.EventService
	logger  *zap.Logger
}

func NewHandler(service interfaces.EventService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var req dto.CreateEventReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	date, err := time.Parse(dateLayout, req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	e, err := h.service.CreateEvent(c.Request.Context(), req.UserID, date, req.Title)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": e})
}

func (h *Handler) UpdateEvent(c *gin.Context) {
	var req dto.UpdateEventReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	date, err := time.Parse(dateLayout, req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	e, err := h.service.UpdateEvent(c.Request.Context(), req.ID, req.UserID, date, req.Title)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": e})
}

func (h *Handler) DeleteEvent(c *gin.Context) {
	var req dto.DeleteEventReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	if err := h.service.DeleteEvent(c.Request.Context(), req.UserID, req.ID); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
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

func (h *Handler) queryEvents(c *gin.Context, kind string) {
	var req dto.QueryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id or date"})
		return
	}
	date, err := time.Parse(dateLayout, req.Date)
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": events})
}
