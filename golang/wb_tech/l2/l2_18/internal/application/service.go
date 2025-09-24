package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/domain/interfaces"
)

type Service struct {
	repo interfaces.EventRepository
}

func NewService(r interfaces.EventRepository) *Service {
	return &Service{repo: r}
}

func (s *Service) CreateEvent(ctx context.Context, userID int64, date time.Time, title string) (entity.Event, error) {
	if title == "" {
		return entity.Event{}, fmt.Errorf("event title empty")
	}
	id := uuid.New().String()
	e := entity.Event{ID: id, UserID: userID, Date: date, Title: title}
	if err := s.repo.Create(ctx, e); err != nil {
		return entity.Event{}, err
	}
	return e, nil
}

func (s *Service) UpdateEvent(ctx context.Context, id string, userID int64, date time.Time, title string) (entity.Event, error) {
	if title == "" {
		return entity.Event{}, fmt.Errorf("event title empty")
	}
	e := entity.Event{ID: id, UserID: userID, Date: date, Title: title}
	if err := s.repo.Update(ctx, e); err != nil {
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
