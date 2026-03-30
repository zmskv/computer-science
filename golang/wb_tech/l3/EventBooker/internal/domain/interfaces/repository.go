package interfaces

import (
	"context"
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/entity"
)

type EventRepository interface {
	CreateUser(ctx context.Context, user entity.User) error
	ListUsers(ctx context.Context) ([]entity.User, error)
	GetUser(ctx context.Context, userID string) (entity.User, error)
	CreateEvent(ctx context.Context, event entity.Event) error
	ListEvents(ctx context.Context) ([]entity.EventDetails, error)
	GetEvent(ctx context.Context, eventID string) (entity.EventDetails, error)
	CreateBooking(ctx context.Context, booking entity.Booking, now time.Time) (entity.Booking, error)
	ConfirmBooking(ctx context.Context, eventID, bookingID string, now time.Time) (entity.Booking, error)
	ExpireBookings(ctx context.Context, now time.Time) ([]entity.ExpiredBookingNotice, error)
	Close()
}
