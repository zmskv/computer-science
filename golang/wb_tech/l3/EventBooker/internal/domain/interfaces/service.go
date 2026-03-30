package interfaces

import (
	"context"
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/entity"
)

type EventService interface {
	RegisterUser(ctx context.Context, name, email, telegramChatID string) (entity.User, error)
	ListUsers(ctx context.Context) ([]entity.User, error)
	CreateEvent(
		ctx context.Context,
		name string,
		startsAt time.Time,
		capacity int,
		requiresConfirmation bool,
		bookingTTLMinutes int,
	) (entity.Event, error)
	ListEvents(ctx context.Context) ([]entity.EventDetails, error)
	GetEvent(ctx context.Context, eventID string) (entity.EventDetails, error)
	BookSeat(ctx context.Context, eventID, userID string) (entity.Booking, error)
	ConfirmBooking(ctx context.Context, eventID, bookingID string) (entity.Booking, error)
	ExpireBookings(ctx context.Context) (int, error)
}
