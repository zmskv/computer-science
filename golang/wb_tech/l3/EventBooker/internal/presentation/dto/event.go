package dto

import (
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/entity"
)

type CreateEventRequest struct {
	Name                 string `json:"name"`
	StartsAt             string `json:"starts_at"`
	Capacity             int    `json:"capacity"`
	RequiresConfirmation bool   `json:"requires_confirmation"`
	BookingTTLMinutes    int    `json:"booking_ttl_minutes"`
}

type CreateUserRequest struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	TelegramChatID string `json:"telegram_chat_id"`
}

type BookSeatRequest struct {
	UserID string `json:"user_id"`
}

type ConfirmBookingRequest struct {
	BookingID string `json:"booking_id"`
}

type BookingResponse struct {
	ID            string               `json:"id"`
	EventID       string               `json:"event_id"`
	UserID        string               `json:"user_id"`
	CustomerName  string               `json:"customer_name"`
	CustomerEmail string               `json:"customer_email,omitempty"`
	Status        entity.BookingStatus `json:"status"`
	ExpiresAt     *time.Time           `json:"expires_at,omitempty"`
	ConfirmedAt   *time.Time           `json:"confirmed_at,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	User          *UserResponse        `json:"user,omitempty"`
}

type UserResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	TelegramChatID string    `json:"telegram_chat_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type EventResponse struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	StartsAt             time.Time         `json:"starts_at"`
	Capacity             int               `json:"capacity"`
	RequiresConfirmation bool              `json:"requires_confirmation"`
	BookingTTLMinutes    int               `json:"booking_ttl_minutes"`
	AvailableSeats       int               `json:"available_seats"`
	PendingBookings      int               `json:"pending_bookings"`
	ConfirmedBookings    int               `json:"confirmed_bookings"`
	Bookings             []BookingResponse `json:"bookings"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

func EventFromEntity(event entity.EventDetails) EventResponse {
	bookings := make([]BookingResponse, 0, len(event.Bookings))
	for _, booking := range event.Bookings {
		bookings = append(bookings, BookingFromEntity(booking))
	}

	return EventResponse{
		ID:                   event.ID,
		Name:                 event.Name,
		StartsAt:             event.StartsAt,
		Capacity:             event.Capacity,
		RequiresConfirmation: event.RequiresConfirmation,
		BookingTTLMinutes:    event.BookingTTLMinutes,
		AvailableSeats:       event.AvailableSeats,
		PendingBookings:      event.PendingBookings,
		ConfirmedBookings:    event.ConfirmedBookings,
		Bookings:             bookings,
		CreatedAt:            event.CreatedAt,
		UpdatedAt:            event.UpdatedAt,
	}
}

func BookingFromEntity(booking entity.Booking) BookingResponse {
	response := BookingResponse{
		ID:            booking.ID,
		EventID:       booking.EventID,
		UserID:        booking.UserID,
		CustomerName:  booking.CustomerName,
		CustomerEmail: booking.CustomerEmail,
		Status:        booking.Status,
		ExpiresAt:     booking.ExpiresAt,
		ConfirmedAt:   booking.ConfirmedAt,
		CreatedAt:     booking.CreatedAt,
		UpdatedAt:     booking.UpdatedAt,
	}

	if booking.User != nil {
		user := UserFromEntity(*booking.User)
		response.User = &user
	}

	return response
}

func UserFromEntity(user entity.User) UserResponse {
	return UserResponse{
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		TelegramChatID: user.TelegramChatID,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}
