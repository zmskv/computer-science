package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/interfaces"
	presentationdto "github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/presentation/dto"
	"go.uber.org/zap"
)

type Handler struct {
	service interfaces.EventService
	logger  *zap.Logger
	webRoot string
}

func NewHandler(service interfaces.EventService, logger *zap.Logger, webRoot string) *Handler {
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

func (h *Handler) CreateUser(c *ginext.Context) {
	var req presentationdto.CreateUserRequest
	if err := decodeJSON(c.Request, &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	user, err := h.service.RegisterUser(c.Request.Context(), req.Name, req.Email, req.TelegramChatID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, presentationdto.UserFromEntity(user))
}

func (h *Handler) ListUsers(c *ginext.Context) {
	users, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		h.logger.Error("request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
		return
	}

	response := make([]presentationdto.UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, presentationdto.UserFromEntity(user))
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) CreateEvent(c *ginext.Context) {
	var req presentationdto.CreateEventRequest
	if err := decodeJSON(c.Request, &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	startsAt, err := parseDateTime(req.StartsAt)
	if err != nil {
		err = errors.Join(application.ErrInvalidInput, err)
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	event, err := h.service.CreateEvent(
		c.Request.Context(),
		req.Name,
		startsAt,
		req.Capacity,
		req.RequiresConfirmation,
		req.BookingTTLMinutes,
	)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
		}
		return
	}

	eventDetails := presentationdto.EventResponse{
		ID:                   event.ID,
		Name:                 event.Name,
		StartsAt:             event.StartsAt,
		Capacity:             event.Capacity,
		RequiresConfirmation: event.RequiresConfirmation,
		BookingTTLMinutes:    event.BookingTTLMinutes,
		AvailableSeats:       event.Capacity,
		PendingBookings:      0,
		ConfirmedBookings:    0,
		Bookings:             []presentationdto.BookingResponse{},
		CreatedAt:            event.CreatedAt,
		UpdatedAt:            event.UpdatedAt,
	}

	c.JSON(http.StatusCreated, eventDetails)
}

func (h *Handler) ListEvents(c *ginext.Context) {
	events, err := h.service.ListEvents(c.Request.Context())
	if err != nil {
		h.logger.Error("request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
		return
	}

	response := make([]presentationdto.EventResponse, 0, len(events))
	for _, event := range events {
		response = append(response, presentationdto.EventFromEntity(event))
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetEvent(c *ginext.Context) {
	eventDetails, err := h.service.GetEvent(c.Request.Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		case errors.Is(err, application.ErrEventNotFound):
			c.JSON(http.StatusNotFound, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, presentationdto.EventFromEntity(eventDetails))
}

func (h *Handler) BookSeat(c *ginext.Context) {
	var req presentationdto.BookSeatRequest
	if err := decodeJSON(c.Request, &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	booking, err := h.service.BookSeat(
		c.Request.Context(),
		c.Param("id"),
		req.UserID,
	)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		case errors.Is(err, application.ErrEventNotFound), errors.Is(err, application.ErrUserNotFound):
			c.JSON(http.StatusNotFound, ginext.H{"error": err.Error()})
		case errors.Is(err, application.ErrNoSeatsAvailable),
			errors.Is(err, application.ErrEventAlreadyStarted):
			c.JSON(http.StatusConflict, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, presentationdto.BookingFromEntity(booking))
}

func (h *Handler) ConfirmBooking(c *ginext.Context) {
	var req presentationdto.ConfirmBookingRequest
	if err := decodeJSON(c.Request, &req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	booking, err := h.service.ConfirmBooking(c.Request.Context(), c.Param("id"), req.BookingID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		case errors.Is(err, application.ErrEventNotFound), errors.Is(err, application.ErrBookingNotFound):
			c.JSON(http.StatusNotFound, ginext.H{"error": err.Error()})
		case errors.Is(err, application.ErrBookingExpired),
			errors.Is(err, application.ErrBookingAlreadyConfirmed),
			errors.Is(err, application.ErrConfirmationNotRequired):
			c.JSON(http.StatusConflict, ginext.H{"error": err.Error()})
		default:
			h.logger.Error("request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, presentationdto.BookingFromEntity(booking))
}

func (h *Handler) serveFrontendFile(c *ginext.Context, filename, contentType string) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Content-Type", contentType)
	c.File(filepath.Join(h.webRoot, filename))
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

func parseDateTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, errors.New("starts_at is required")
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02 15:04",
	}

	for _, layout := range layouts {
		if layout == time.RFC3339 {
			parsed, err := time.Parse(layout, trimmed)
			if err == nil {
				return parsed.UTC(), nil
			}
			continue
		}

		parsed, err := time.ParseInLocation(layout, trimmed, time.Local)
		if err == nil {
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, errors.New("starts_at must be a valid date-time")
}
