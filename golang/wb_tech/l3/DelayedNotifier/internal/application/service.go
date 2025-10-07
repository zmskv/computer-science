package application

import (
	"context"
	"fmt"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/interfaces"
	"go.uber.org/zap"
)

type NotifierService struct {
	repo        interfaces.NotifierRepository
	emailClient interfaces.SMTPClient
	logger      *zap.Logger
}

func NewNotifierService(repo interfaces.NotifierRepository, emailClient interfaces.SMTPClient, logger *zap.Logger) interfaces.NotifierService {
	return &NotifierService{repo: repo, emailClient: emailClient, logger: logger}
}

func (s *NotifierService) CreateNote(ctx context.Context, note entity.Note) (string, error) {
	id, err := s.repo.CreateNote(ctx, note)
	if err != nil {
		s.logger.Error("failed to create and schedule note", zap.Error(err))
		return "", err
	}
	s.logger.Info("Note successfully scheduled in Redis", zap.String("id", id))
	return id, nil
}

func (s *NotifierService) GetNote(ctx context.Context, id string) (entity.Note, error) {
	return s.repo.GetNote(ctx, id)
}

func (s *NotifierService) DeleteNote(ctx context.Context, id string) error {
	return s.repo.DeleteNote(ctx, id)
}

func (s *NotifierService) UpdateNoteStatus(ctx context.Context, id, status string) error {
	return s.repo.UpdateNoteStatus(ctx, id, status)
}

func (s *NotifierService) UpdateNoteRetries(ctx context.Context, id string, retries int) error {
	return s.repo.UpdateNoteRetries(ctx, id, retries)
}

func (s *NotifierService) SendNotification(ctx context.Context, note entity.Note) error {
	if note.Channel == "email" {
		s.logger.Info("Would send email", zap.String("to", note.Recipient), zap.String("subject", note.Title))
		return s.emailClient.SendEmail(note.Recipient, note.Title, note.Body)
	}
	return fmt.Errorf("unknown channel: %s", note.Channel)
}
