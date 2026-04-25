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
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/interfaces"
	presentationdto "github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/presentation/dto"
	presentationmiddleware "github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/presentation/middleware"
	presentationresponses "github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/presentation/responses"
	"go.uber.org/zap"
)

type Handler struct {
	warehouseService interfaces.WarehouseService
	authService      interfaces.AuthService
	logger           *zap.Logger
	webRoot          string
}

func NewHandler(
	warehouseService interfaces.WarehouseService,
	authService interfaces.AuthService,
	logger *zap.Logger,
	webRoot string,
) *Handler {
	return &Handler{
		warehouseService: warehouseService,
		authService:      authService,
		logger:           logger,
		webRoot:          webRoot,
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

func (h *Handler) Login(c *ginext.Context) {
	var req presentationdto.LoginRequest
	if err := decodeJSON(c.Request, &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	session, err := h.authService.Login(c.Request.Context(), req.Username, entity.Role(strings.ToLower(strings.TrimSpace(req.Role))))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("login failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, presentationresponses.AuthSessionFromEntity(session))
}

func (h *Handler) CreateItem(c *ginext.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}

	var req presentationdto.CreateItemRequest
	if err := decodeJSON(c.Request, &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	item, err := h.warehouseService.CreateItem(c.Request.Context(), actor, entity.ItemMutation{
		Name:        req.Name,
		SKU:         req.SKU,
		Quantity:    req.Quantity,
		Location:    req.Location,
		Description: req.Description,
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, presentationresponses.ItemFromEntity(item))
}

func (h *Handler) GetItem(c *ginext.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}

	item, err := h.warehouseService.GetItem(c.Request.Context(), actor, c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationresponses.ItemFromEntity(item))
}

func (h *Handler) ListItems(c *ginext.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}

	items, err := h.warehouseService.ListItems(c.Request.Context(), actor, entity.ItemFilter{
		Query: c.Query("q"),
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationresponses.ItemsFromEntities(items))
}

func (h *Handler) UpdateItem(c *ginext.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}

	var req presentationdto.UpdateItemRequest
	if err := decodeJSON(c.Request, &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	item, err := h.warehouseService.UpdateItem(c.Request.Context(), actor, c.Param("id"), entity.ItemMutation{
		Name:        req.Name,
		SKU:         req.SKU,
		Quantity:    req.Quantity,
		Location:    req.Location,
		Description: req.Description,
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationresponses.ItemFromEntity(item))
}

func (h *Handler) DeleteItem(c *ginext.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}

	if err := h.warehouseService.DeleteItem(c.Request.Context(), actor, c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListItemHistory(c *ginext.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}

	filter, err := buildHistoryFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}
	filter.ItemID = c.Param("id")

	entries, err := h.warehouseService.ListHistory(c.Request.Context(), actor, filter)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationresponses.HistoryEntriesFromEntities(entries))
}

func (h *Handler) ListHistory(c *ginext.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}

	filter, err := buildHistoryFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	entries, err := h.warehouseService.ListHistory(c.Request.Context(), actor, filter)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationresponses.HistoryEntriesFromEntities(entries))
}

func (h *Handler) ExportHistoryCSV(c *ginext.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}

	filter, err := buildHistoryFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	entries, err := h.warehouseService.ListHistory(c.Request.Context(), actor, filter)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	rows := [][]string{
		{"id", "item_id", "action", "changed_by", "changed_role", "changed_at", "changes"},
	}
	for _, entry := range entries {
		rows = append(rows, []string{
			fmt.Sprintf("%d", entry.ID),
			entry.ItemID,
			string(entry.Action),
			entry.ChangedBy,
			string(entry.ChangedRole),
			entry.ChangedAt.Format(time.RFC3339),
			formatHistoryChanges(entry.Changes),
		})
	}

	if err := writer.WriteAll(rows); err != nil {
		h.logger.Error("failed to build csv", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
		return
	}

	filename := fmt.Sprintf("warehouse-history-%s.csv", time.Now().UTC().Format("20060102-150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

func (h *Handler) actorFromContext(c *ginext.Context) (entity.Actor, bool) {
	actor, err := presentationmiddleware.ActorFromContext(c)
	if err != nil {
		h.logger.Error("authenticated actor missing in context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
		return entity.Actor{}, false
	}

	return actor, true
}

func (h *Handler) writeServiceError(c *ginext.Context, err error) {
	switch {
	case errors.Is(err, application.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
	case errors.Is(err, application.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, ginext.H{"error": err.Error()})
	case errors.Is(err, application.ErrForbidden):
		c.JSON(http.StatusForbidden, ginext.H{"error": err.Error()})
	case errors.Is(err, entity.ErrItemNotFound):
		c.JSON(http.StatusNotFound, ginext.H{"error": err.Error()})
	default:
		h.logger.Error("request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
	}
}

func (h *Handler) serveFrontendFile(c *ginext.Context, filename, contentType string) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Content-Type", contentType)
	c.File(filepath.Join(h.webRoot, filename))
}

func buildHistoryFilter(c *ginext.Context) (entity.HistoryFilter, error) {
	from, err := parseOptionalTime(c.Query("from"), false)
	if err != nil {
		return entity.HistoryFilter{}, fmt.Errorf("%w: from must be a valid date or date-time", application.ErrInvalidInput)
	}

	to, err := parseOptionalTime(c.Query("to"), true)
	if err != nil {
		return entity.HistoryFilter{}, fmt.Errorf("%w: to must be a valid date or date-time", application.ErrInvalidInput)
	}

	return entity.HistoryFilter{
		ItemID:   c.Query("item_id"),
		Username: c.Query("username"),
		Action:   entity.HistoryAction(strings.ToLower(strings.TrimSpace(c.Query("action")))),
		From:     from,
		To:       to,
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

func formatHistoryChanges(changes []entity.HistoryChange) string {
	if len(changes) == 0 {
		return ""
	}

	parts := make([]string, 0, len(changes))
	for _, change := range changes {
		parts = append(parts, fmt.Sprintf("%s: %s -> %s", change.Field, change.Before, change.After))
	}

	return strings.Join(parts, "; ")
}
