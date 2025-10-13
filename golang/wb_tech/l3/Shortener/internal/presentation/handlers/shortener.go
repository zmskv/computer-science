package handlers

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/domain/interfaces"
	"go.uber.org/zap"
)

type Handler struct {
	service interfaces.ShortenerService
	logger  *zap.Logger
}

func NewHandler(service interfaces.ShortenerService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Shorten(c *ginext.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid request format"})
		return
	}

	su, err := h.service.Create(c.Request.Context(), req.URL)
	if err != nil {
		h.logger.Error("failed to create short URL", zap.Error(err), zap.String("url", req.URL))
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	h.logger.Info("short URL created", zap.String("short_code", su.ShortCode), zap.String("original_url", su.OriginalURL))
	c.JSON(http.StatusCreated, su)
}

func (h *Handler) Redirect(c *ginext.Context) {
	code := c.Param("short_url")
	url, err := h.service.Resolve(c.Request.Context(), code, c.GetHeader("User-Agent"))
	if err != nil {
		h.logger.Error("failed to resolve short URL", zap.Error(err), zap.String("short_code", code))
		c.JSON(http.StatusNotFound, ginext.H{"error": "short URL not found"})
		return
	}

	h.logger.Info("redirecting", zap.String("short_code", code), zap.String("target_url", url))
	c.Redirect(http.StatusFound, url)
}

func (h *Handler) Analytics(c *ginext.Context) {
	code := c.Param("short_url")
	data, err := h.service.Analytics(c.Request.Context(), code)
	if err != nil {
		h.logger.Error("failed to get analytics", zap.Error(err), zap.String("short_code", code))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to get analytics"})
		return
	}

	h.logger.Info("analytics requested", zap.String("short_code", code))
	c.JSON(http.StatusOK, data)
}
