package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/interfaces"
	"go.uber.org/zap"
)

var (
	ErrInvalidInput            = errors.New("invalid input")
	ErrEventNotFound           = errors.New("event not found")
	ErrBookingNotFound         = errors.New("booking not found")
	ErrUserNotFound            = errors.New("user not found")
	ErrNoSeatsAvailable        = errors.New("no seats available")
	ErrBookingExpired          = errors.New("booking expired")
	ErrBookingAlreadyConfirmed = errors.New("booking already confirmed")
	ErrConfirmationNotRequired = errors.New("confirmation is not required for this event")
	ErrEventAlreadyStarted     = errors.New("event already started")
)

type eventService struct {
	repo              interfaces.EventRepository
	defaultBookingTTL time.Duration
	notifier          interfaces.BookingNotifier
	logger            *zap.Logger
}

func NewEventService(
	repo interfaces.EventRepository,
	defaultBookingTTL time.Duration,
	notifier interfaces.BookingNotifier,
	logger *zap.Logger,
) interfaces.EventService {
	return &eventService{
		repo:              repo,
		defaultBookingTTL: defaultBookingTTL,
		notifier:          notifier,
		logger:            logger,
	}
}

func (s *eventService) RegisterUser(ctx context.Context, name, email, telegramChatID string) (entity.User, error) {
	now := time.Now().UTC()
	user := entity.User{
		ID:             uuid.NewString(),
		Name:           strings.TrimSpace(name),
		Email:          strings.TrimSpace(email),
		TelegramChatID: strings.TrimSpace(telegramChatID),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := user.Validate(); err != nil {
		return entity.User{}, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (s *eventService) ListUsers(ctx context.Context) ([]entity.User, error) {
	return s.repo.ListUsers(ctx)
}

func (s *eventService) CreateEvent(
	ctx context.Context,
	name string,
	startsAt time.Time,
	capacity int,
	requiresConfirmation bool,
	bookingTTLMinutes int,
) (entity.Event, error) {
	trimmedName := strings.TrimSpace(name)

	if requiresConfirmation && bookingTTLMinutes == 0 {
		bookingTTLMinutes = int(s.defaultBookingTTL / time.Minute)
	}

	now := time.Now().UTC()

	event := entity.Event{
		ID:                   uuid.NewString(),
		Name:                 trimmedName,
		StartsAt:             startsAt.UTC(),
		Capacity:             capacity,
		RequiresConfirmation: requiresConfirmation,
		BookingTTLMinutes:    bookingTTLMinutes,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := event.Validate(); err != nil {
		return entity.Event{}, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return entity.Event{}, err
	}

	return event, nil
}

func (s *eventService) ListEvents(ctx context.Context) ([]entity.EventDetails, error) {
	return s.repo.ListEvents(ctx)
}

func (s *eventService) GetEvent(ctx context.Context, eventID string) (entity.EventDetails, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return entity.EventDetails{}, fmt.Errorf("%w: event id is required", ErrInvalidInput)
	}

	return s.repo.GetEvent(ctx, eventID)
}

func (s *eventService) BookSeat(ctx context.Context, eventID, userID string) (entity.Booking, error) {
	eventID = strings.TrimSpace(eventID)
	userID = strings.TrimSpace(userID)

	if eventID == "" {
		return entity.Booking{}, fmt.Errorf("%w: event id is required", ErrInvalidInput)
	}

	if userID == "" {
		return entity.Booking{}, fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return entity.Booking{}, err
	}

	eventDetails, err := s.repo.GetEvent(ctx, eventID)
	if err != nil {
		return entity.Booking{}, err
	}

	now := time.Now().UTC()
	if !eventDetails.StartsAt.After(now) {
		return entity.Booking{}, ErrEventAlreadyStarted
	}

	booking := entity.Booking{
		ID:            uuid.NewString(),
		EventID:       eventID,
		UserID:        user.ID,
		CustomerName:  user.Name,
		CustomerEmail: user.Email,
		CreatedAt:     now,
		UpdatedAt:     now,
		User:          &user,
	}

	if eventDetails.RequiresConfirmation {
		expiresAt := now.Add(time.Duration(eventDetails.BookingTTLMinutes) * time.Minute)
		booking.Status = entity.BookingStatusPending
		booking.ExpiresAt = &expiresAt
	} else {
		confirmedAt := now
		booking.Status = entity.BookingStatusConfirmed
		booking.ConfirmedAt = &confirmedAt
	}

	if err := booking.Validate(); err != nil {
		return entity.Booking{}, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	return s.repo.CreateBooking(ctx, booking, now)
}

func (s *eventService) ConfirmBooking(ctx context.Context, eventID, bookingID string) (entity.Booking, error) {
	eventID = strings.TrimSpace(eventID)
	bookingID = strings.TrimSpace(bookingID)

	if eventID == "" || bookingID == "" {
		return entity.Booking{}, fmt.Errorf("%w: event id and booking id are required", ErrInvalidInput)
	}

	return s.repo.ConfirmBooking(ctx, eventID, bookingID, time.Now().UTC())
}

func (s *eventService) ExpireBookings(ctx context.Context) (int, error) {
	expired, err := s.repo.ExpireBookings(ctx, time.Now().UTC())
	if err != nil {
		s.logger.Error("failed to expire bookings", zap.Error(err))
		return 0, err
	}

	for _, notice := range expired {
		if s.notifier == nil {
			continue
		}

		if err := s.notifier.NotifyBookingExpired(ctx, notice); err != nil {
			s.logger.Error(
				"failed to notify about expired booking",
				zap.String("booking_id", notice.Booking.ID),
				zap.String("user_id", notice.User.ID),
				zap.Error(err),
			)
		}
	}

	return len(expired), nil
}
