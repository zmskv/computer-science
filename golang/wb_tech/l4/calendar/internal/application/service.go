package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"calendar/internal/domain/entity"
	"calendar/internal/domain/interfaces"
)

type Config struct {
	ReminderBuffer  int
	ArchiveInterval time.Duration
	ArchiveAfter    time.Duration
}

type reminderTask struct {
	EventID    string
	ReminderAt time.Time
}

type Service struct {
	repo            interfaces.EventRepository
	notifier        interfaces.ReminderSender
	logger          interfaces.Logger
	reminderTasks   chan reminderTask
	archiveInterval time.Duration
	archiveAfter    time.Duration
}

func NewService(
	r interfaces.EventRepository,
	notifier interfaces.ReminderSender,
	logger interfaces.Logger,
	cfg Config,
) *Service {
	if cfg.ReminderBuffer <= 0 {
		cfg.ReminderBuffer = 128
	}
	if cfg.ArchiveInterval <= 0 {
		cfg.ArchiveInterval = time.Minute
	}
	if cfg.ArchiveAfter <= 0 {
		cfg.ArchiveAfter = 24 * time.Hour
	}
	if notifier == nil {
		notifier = noopReminderSender{}
	}
	if logger == nil {
		logger = noopLogger{}
	}

	return &Service{
		repo:            r,
		notifier:        notifier,
		logger:          logger,
		reminderTasks:   make(chan reminderTask, cfg.ReminderBuffer),
		archiveInterval: cfg.ArchiveInterval,
		archiveAfter:    cfg.ArchiveAfter,
	}
}

func (s *Service) Start(ctx context.Context) {
	go s.runReminderDispatcher(ctx)
	go s.runArchiveLoop(ctx)
}

func (s *Service) CreateEvent(
	ctx context.Context,
	userID int64,
	date time.Time,
	title string,
	reminderAt *time.Time,
) (entity.Event, error) {
	if strings.TrimSpace(title) == "" {
		return entity.Event{}, fmt.Errorf("event title empty")
	}
	if userID <= 0 {
		return entity.Event{}, fmt.Errorf("user_id must be positive")
	}

	id, err := newID()
	if err != nil {
		return entity.Event{}, fmt.Errorf("generate id: %w", err)
	}

	e := entity.Event{
		ID:         id,
		UserID:     userID,
		Date:       date.UTC(),
		Title:      strings.TrimSpace(title),
		ReminderAt: cloneTime(reminderAt),
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return entity.Event{}, err
	}

	if err := s.enqueueReminder(ctx, e); err != nil {
		return entity.Event{}, err
	}

	return e, nil
}

func (s *Service) UpdateEvent(
	ctx context.Context,
	id string,
	userID int64,
	date time.Time,
	title string,
	reminderAt *time.Time,
) (entity.Event, error) {
	if strings.TrimSpace(title) == "" {
		return entity.Event{}, fmt.Errorf("event title empty")
	}
	if id == "" {
		return entity.Event{}, fmt.Errorf("id empty")
	}
	if userID <= 0 {
		return entity.Event{}, fmt.Errorf("user_id must be positive")
	}

	e := entity.Event{
		ID:           id,
		UserID:       userID,
		Date:         date.UTC(),
		Title:        strings.TrimSpace(title),
		ReminderAt:   cloneTime(reminderAt),
		ReminderSent: false,
	}
	if err := s.repo.Update(ctx, e); err != nil {
		return entity.Event{}, err
	}

	if err := s.enqueueReminder(ctx, e); err != nil {
		return entity.Event{}, err
	}

	return e, nil
}

func (s *Service) DeleteEvent(ctx context.Context, userID int64, id string) error {
	if id == "" {
		return fmt.Errorf("id empty")
	}
	return s.repo.Delete(ctx, userID, id)
}

func (s *Service) EventsForDay(ctx context.Context, userID int64, date time.Time) ([]entity.Event, error) {
	return s.repo.FindByDay(ctx, userID, date)
}

func (s *Service) EventsForWeek(ctx context.Context, userID int64, date time.Time) ([]entity.Event, error) {
	return s.repo.FindByWeek(ctx, userID, date)
}

func (s *Service) EventsForMonth(ctx context.Context, userID int64, date time.Time) ([]entity.Event, error) {
	return s.repo.FindByMonth(ctx, userID, date)
}

func (s *Service) ArchivedEvents(ctx context.Context, userID int64) ([]entity.Event, error) {
	return s.repo.ListArchived(ctx, userID)
}

func (s *Service) enqueueReminder(ctx context.Context, event entity.Event) error {
	if event.ReminderAt == nil {
		return nil
	}

	task := reminderTask{
		EventID:    event.ID,
		ReminderAt: event.ReminderAt.UTC(),
	}

	select {
	case s.reminderTasks <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) runReminderDispatcher(ctx context.Context) {
	for {
		select {
		case task := <-s.reminderTasks:
			go s.waitAndSendReminder(ctx, task)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) waitAndSendReminder(ctx context.Context, task reminderTask) {
	wait := time.Until(task.ReminderAt)
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}
	}

	s.sendReminder(ctx, task.EventID)
}

func (s *Service) sendReminder(ctx context.Context, eventID string) {
	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		s.logger.Error("reminder lookup failed", map[string]any{
			"event_id": eventID,
			"error":    err.Error(),
		})
		return
	}

	if event.Archived || event.ReminderAt == nil || event.ReminderSent {
		return
	}
	if time.Until(*event.ReminderAt) > 0 {
		return
	}

	if err := s.notifier.Send(ctx, event); err != nil {
		s.logger.Error("reminder delivery failed", map[string]any{
			"event_id": event.ID,
			"error":    err.Error(),
		})
		return
	}

	if err := s.repo.MarkReminderSent(ctx, event.ID); err != nil {
		s.logger.Error("mark reminder as sent failed", map[string]any{
			"event_id": event.ID,
			"error":    err.Error(),
		})
		return
	}

	s.logger.Info("reminder sent", map[string]any{
		"event_id": event.ID,
		"user_id":  event.UserID,
	})
}

func (s *Service) runArchiveLoop(ctx context.Context) {
	ticker := time.NewTicker(s.archiveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			archived, err := s.repo.ArchiveBefore(ctx, time.Now().UTC().Add(-s.archiveAfter))
			if err != nil {
				s.logger.Error("archive sweep failed", map[string]any{"error": err.Error()})
				continue
			}
			if len(archived) > 0 {
				s.logger.Info("archived old events", map[string]any{"count": len(archived)})
			}
		case <-ctx.Done():
			return
		}
	}
}

func newID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func cloneTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	copy := t.UTC()
	return &copy
}

type noopReminderSender struct{}

func (noopReminderSender) Send(context.Context, entity.Event) error {
	return nil
}

type noopLogger struct{}

func (noopLogger) Info(string, map[string]any)  {}
func (noopLogger) Error(string, map[string]any) {}
