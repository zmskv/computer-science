package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/domain/interfaces"
	presentationdto "github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/presentation/dto"
	presentationresponses "github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/presentation/responses"
	"go.uber.org/zap"
)

type Handler struct {
	service interfaces.SalesService
	logger  *zap.Logger
	webRoot string
}

func NewHandler(service interfaces.SalesService, logger *zap.Logger, webRoot string) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
		webRoot: webRoot,
	}
}

func (h *Handler) Index(c *ginext.Context) {
	h.serveFrontendFile(c, "index.html", "text/html; charset=utf-8")
}

func (h *Handler) AppJS(c *ginext.Context) {
	h.serveFrontendFile(c, "app.js", "application/javascript; charset=utf-8")
}

func (h *Handler) StylesCSS(c *ginext.Context) {
	h.serveFrontendFile(c, "styles.css", "text/css; charset=utf-8")
}

func (h *Handler) Health(c *ginext.Context) {
	c.JSON(http.StatusOK, ginext.H{"status": "ok"})
}

func (h *Handler) CreateItem(c *ginext.Context) {
	var req presentationdto.CreateItemRequest
	if err := decodeJSON(c.Request, &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	occurredAt, err := parseRequiredTime(req.OccurredAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": errors.Join(application.ErrInvalidInput, err).Error()})
		return
	}

	item, err := h.service.CreateItem(
		c.Request.Context(),
		entity.ItemType(strings.ToLower(strings.TrimSpace(req.Type))),
		req.Amount,
		req.Category,
		req.Description,
		occurredAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusCreated, presentationresponses.ItemFromEntity(item))
}

func (h *Handler) GetItem(c *ginext.Context) {
	item, err := h.service.GetItem(c.Request.Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		case errors.Is(err, application.ErrItemNotFound):
			c.JSON(http.StatusNotFound, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, presentationresponses.ItemFromEntity(item))
}

func (h *Handler) ListItems(c *ginext.Context) {
	filter, err := buildItemFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	items, err := h.service.ListItems(c.Request.Context(), filter)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, presentationresponses.ItemsFromEntities(items))
}

func (h *Handler) UpdateItem(c *ginext.Context) {
	var req presentationdto.UpdateItemRequest
	if err := decodeJSON(c.Request, &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	occurredAt, err := parseRequiredTime(req.OccurredAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": errors.Join(application.ErrInvalidInput, err).Error()})
		return
	}

	item, err := h.service.UpdateItem(
		c.Request.Context(),
		c.Param("id"),
		entity.ItemType(strings.ToLower(strings.TrimSpace(req.Type))),
		req.Amount,
		req.Category,
		req.Description,
		occurredAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		case errors.Is(err, application.ErrItemNotFound):
			c.JSON(http.StatusNotFound, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, presentationresponses.ItemFromEntity(item))
}

func (h *Handler) DeleteItem(c *ginext.Context) {
	if err := h.service.DeleteItem(c.Request.Context(), c.Param("id")); err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		case errors.Is(err, application.ErrItemNotFound):
			c.JSON(http.StatusNotFound, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "внутренняя ошибка сервера"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetAnalytics(c *ginext.Context) {
	filter, err := buildAnalyticsFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	result, err := h.service.GetAnalytics(c.Request.Context(), filter)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "внутренняя ошибка сервера"})
		}
		return
	}

	c.JSON(http.StatusOK, presentationresponses.AnalyticsFromEntity(result))
}

func (h *Handler) ExportItemsCSV(c *ginext.Context) {
	filter, err := buildItemFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	items, err := h.service.ListItems(c.Request.Context(), filter)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "внутренняя ошибка сервера"})
		}
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	rows := [][]string{
		{"id", "type", "amount", "category", "description", "occurred_at", "created_at", "updated_at"},
	}
	for _, item := range items {
		rows = append(rows, []string{
			item.ID,
			string(item.Type),
			fmt.Sprintf("%.2f", item.Amount),
			item.Category,
			item.Description,
			item.OccurredAt.Format(time.RFC3339),
			item.CreatedAt.Format(time.RFC3339),
			item.UpdatedAt.Format(time.RFC3339),
		})
	}

	if err := writer.WriteAll(rows); err != nil {
		h.logger.Error("failed to build csv", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "внутренняя ошибка сервера"})
		return
	}

	filename := fmt.Sprintf("sales-items-%s.csv", time.Now().UTC().Format("20060102-150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

func (h *Handler) serveFrontendFile(c *ginext.Context, filename, contentType string) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Content-Type", contentType)
	c.File(filepath.Join(h.webRoot, filename))
}

func buildItemFilter(c *ginext.Context) (entity.ItemFilter, error) {
	from, err := parseOptionalTime(c.Query("from"), false)
	if err != nil {
		return entity.ItemFilter{}, errors.Join(application.ErrInvalidInput, fmt.Errorf("from: %w", err))
	}

	to, err := parseOptionalTime(c.Query("to"), true)
	if err != nil {
		return entity.ItemFilter{}, errors.Join(application.ErrInvalidInput, fmt.Errorf("to: %w", err))
	}

	return entity.ItemFilter{
		From:      from,
		To:        to,
		Type:      entity.ItemType(strings.ToLower(strings.TrimSpace(c.Query("type")))),
		Category:  c.Query("category"),
		SortBy:    entity.SortField(strings.ToLower(strings.TrimSpace(c.Query("sort_by")))),
		SortOrder: entity.SortDirection(strings.ToLower(strings.TrimSpace(c.Query("sort_order")))),
	}, nil
}

func buildAnalyticsFilter(c *ginext.Context) (entity.AnalyticsFilter, error) {
	from, err := parseOptionalTime(c.Query("from"), false)
	if err != nil {
		return entity.AnalyticsFilter{}, errors.Join(application.ErrInvalidInput, fmt.Errorf("from: %w", err))
	}

	to, err := parseOptionalTime(c.Query("to"), true)
	if err != nil {
		return entity.AnalyticsFilter{}, errors.Join(application.ErrInvalidInput, fmt.Errorf("to: %w", err))
	}

	return entity.AnalyticsFilter{
		From:     from,
		To:       to,
		Type:     entity.ItemType(strings.ToLower(strings.TrimSpace(c.Query("type")))),
		Category: c.Query("category"),
		GroupBy:  entity.AnalyticsGroup(strings.ToLower(strings.TrimSpace(c.Query("group_by")))),
	}, nil
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return errors.Join(application.ErrInvalidInput, err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%w: request body must contain a single JSON object", application.ErrInvalidInput)
	}

	return nil
}

func parseRequiredTime(value string) (time.Time, error) {
	parsed, err := parseOptionalTime(value, false)
	if err != nil {
		return time.Time{}, err
	}
	if parsed == nil {
		return time.Time{}, errors.New("occurred_at is required")
	}

	return *parsed, nil
}

func parseOptionalTime(value string, endOfDay bool) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	dateOnlyLayouts := map[string]bool{
		"2006-01-02": true,
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if layout == time.RFC3339 {
			parsed, err := time.Parse(layout, trimmed)
			if err == nil {
				value := parsed.UTC()
				return &value, nil
			}
			continue
		}

		parsed, err := time.ParseInLocation(layout, trimmed, time.Local)
		if err != nil {
			continue
		}

		if dateOnlyLayouts[layout] && endOfDay {
			parsed = parsed.Add(24*time.Hour - time.Nanosecond)
		}

		value := parsed.UTC()
		return &value, nil
	}

	return nil, errors.New("must be a valid date or date-time")
}
