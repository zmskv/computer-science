package application

import (
	"context"
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/interfaces"
	"go.uber.org/zap"
)

type ExpirationWorker struct {
	service  interfaces.EventService
	interval time.Duration
	logger   *zap.Logger
}

func NewExpirationWorker(service interfaces.EventService, interval time.Duration, logger *zap.Logger) *ExpirationWorker {
	if interval <= 0 {
		interval = 5 * time.Second
	}

	return &ExpirationWorker{
		service:  service,
		interval: interval,
		logger:   logger,
	}
}

func (w *ExpirationWorker) Run(ctx context.Context) error {
	w.runOnce(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("expiration worker stopped")
			return nil
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *ExpirationWorker) runOnce(ctx context.Context) {
	expired, err := w.service.ExpireBookings(ctx)
	if err != nil {
		w.logger.Error("expiration cycle failed", zap.Error(err))
		return
	}

	if expired > 0 {
		w.logger.Info("expired stale bookings", zap.Int("count", expired))
	}
}
