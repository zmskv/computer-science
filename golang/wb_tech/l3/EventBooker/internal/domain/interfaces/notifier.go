package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/entity"
)

type BookingNotifier interface {
	NotifyBookingExpired(ctx context.Context, notice entity.ExpiredBookingNotice) error
}
