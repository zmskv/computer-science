package entity

import (
	"errors"
	"strings"
	"time"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusExpired   BookingStatus = "expired"
)

type Event struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	StartsAt             time.Time `json:"starts_at"`
	Capacity             int       `json:"capacity"`
	RequiresConfirmation bool      `json:"requires_confirmation"`
	BookingTTLMinutes    int       `json:"booking_ttl_minutes"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type User struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	TelegramChatID string    `json:"telegram_chat_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Booking struct {
	ID            string        `json:"id"`
	EventID       string        `json:"event_id"`
	UserID        string        `json:"user_id"`
	CustomerName  string        `json:"customer_name"`
	CustomerEmail string        `json:"customer_email,omitempty"`
	Status        BookingStatus `json:"status"`
	ExpiresAt     *time.Time    `json:"expires_at,omitempty"`
	ConfirmedAt   *time.Time    `json:"confirmed_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	User          *User         `json:"user,omitempty"`
}

type EventDetails struct {
	Event
	AvailableSeats    int       `json:"available_seats"`
	PendingBookings   int       `json:"pending_bookings"`
	ConfirmedBookings int       `json:"confirmed_bookings"`
	Bookings          []Booking `json:"bookings"`
}

type ExpiredBookingNotice struct {
	Booking Booking `json:"booking"`
	Event   Event   `json:"event"`
	User    User    `json:"user"`
}

func (e Event) Validate() error {
	validators := []func() error{
		e.validateID,
		e.validateName,
		e.validateStartsAt,
		e.validateCapacity,
		e.validateBookingTTL,
	}

	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

func (b Booking) Validate() error {
	validators := []func() error{
		b.validateID,
		b.validateEventID,
		b.validateUserID,
		b.validateCustomerName,
		b.validateStatus,
		b.validateStatusState,
	}

	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

func (u User) Validate() error {
	validators := []func() error{
		u.validateID,
		u.validateName,
		u.validateEmail,
	}

	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

func (s BookingStatus) IsValid() bool {
	switch s {
	case BookingStatusPending, BookingStatusConfirmed, BookingStatusExpired:
		return true
	default:
		return false
	}
}

func (e Event) validateID() error {
	if strings.TrimSpace(e.ID) == "" {
		return errors.New("event id is required")
	}

	return nil
}

func (e Event) validateName() error {
	if strings.TrimSpace(e.Name) == "" {
		return errors.New("event name is required")
	}

	return nil
}

func (e Event) validateStartsAt() error {
	if e.StartsAt.IsZero() {
		return errors.New("event start time is required")
	}

	return nil
}

func (e Event) validateCapacity() error {
	if e.Capacity <= 0 {
		return errors.New("capacity must be greater than zero")
	}

	return nil
}

func (e Event) validateBookingTTL() error {
	if e.BookingTTLMinutes < 0 {
		return errors.New("booking ttl must not be negative")
	}

	return nil
}

func (b Booking) validateID() error {
	if strings.TrimSpace(b.ID) == "" {
		return errors.New("booking id is required")
	}

	return nil
}

func (b Booking) validateEventID() error {
	if strings.TrimSpace(b.EventID) == "" {
		return errors.New("event id is required")
	}

	return nil
}

func (b Booking) validateUserID() error {
	if strings.TrimSpace(b.UserID) == "" {
		return errors.New("user id is required")
	}

	return nil
}

func (b Booking) validateCustomerName() error {
	if strings.TrimSpace(b.CustomerName) == "" {
		return errors.New("customer name is required")
	}

	return nil
}

func (b Booking) validateStatus() error {
	if !b.Status.IsValid() {
		return errors.New("booking status is invalid")
	}

	return nil
}

func (b Booking) validateStatusState() error {
	switch b.Status {
	case BookingStatusPending:
		return b.validatePendingState()
	case BookingStatusConfirmed:
		return b.validateConfirmedState()
	default:
		return nil
	}
}

func (b Booking) validatePendingState() error {
	if b.ExpiresAt == nil {
		return errors.New("pending booking must have expiration time")
	}

	if b.ConfirmedAt != nil {
		return errors.New("pending booking cannot be confirmed")
	}

	return nil
}

func (b Booking) validateConfirmedState() error {
	if b.ConfirmedAt == nil {
		return errors.New("confirmed booking must have confirmation time")
	}

	return nil
}

func (u User) validateID() error {
	if strings.TrimSpace(u.ID) == "" {
		return errors.New("user id is required")
	}

	return nil
}

func (u User) validateName() error {
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("user name is required")
	}

	return nil
}

func (u User) validateEmail() error {
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("user email is required")
	}

	return nil
}
