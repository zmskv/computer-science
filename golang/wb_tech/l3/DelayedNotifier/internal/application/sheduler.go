package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wb-go/wbf/rabbitmq"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/interfaces"
	"go.uber.org/zap"
)

type Scheduler struct {
	repo     interfaces.NotifierRepository
	producer *rabbitmq.Publisher
	logger   *zap.Logger
}

func NewScheduler(repo interfaces.NotifierRepository, producer *rabbitmq.Publisher, logger *zap.Logger) *Scheduler {
	return &Scheduler{repo: repo, producer: producer, logger: logger}
}

func (s *Scheduler) Run(ctx context.Context) {
	s.logger.Info("Scheduler started")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	s.processDueNotifications(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Scheduler stopping...")
			return
		case <-ticker.C:
			s.processDueNotifications(ctx)
		}
	}
}

func (s *Scheduler) processDueNotifications(ctx context.Context) {
	dueNotes, err := s.repo.GetDueNotificationIDs(ctx)
	if err != nil {
		s.logger.Error("Failed to get due notifications from repo", zap.Error(err))
		return
	}

	if len(dueNotes) == 0 {
		return
	}

	s.logger.Info("Found due notifications, preparing to publish", zap.Int("count", len(dueNotes)))

	for _, noteID := range dueNotes {
		s.logger.Info("Publishing note", zap.String("id", noteID))

		if err := s.repo.UpdateNoteStatus(ctx, noteID, "in_queue"); err != nil {
			s.logger.Error("Failed to update note status to in_queue", zap.String("id", noteID), zap.Error(err))
			continue
		}

		payload, _ := json.Marshal(map[string]string{"id": noteID})
		if err := s.producer.Publish(payload, "", "application/json"); err != nil {
			s.logger.Error("Failed to publish note ID to queue", zap.String("id", noteID), zap.Error(err))
			_ = s.repo.UpdateNoteStatus(ctx, noteID, "pending")
		}
	}
}
