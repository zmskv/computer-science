package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/domain/interfaces"
)

var (
	ErrNotFound error = errors.New("event not found")
)

type inMemoryRepo struct {
	mu     sync.RWMutex
	events map[string]entity.Event
}

func NewInMemoryRepo() interfaces.EventRepository {
	return &inMemoryRepo{events: make(map[string]entity.Event)}
}

func (r *inMemoryRepo) Create(ctx context.Context, e entity.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.events[e.ID]; ok {
		return fmt.Errorf("event already exists")
	}
	r.events[e.ID] = e
	return nil
}

func (r *inMemoryRepo) Update(ctx context.Context, e entity.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.events[e.ID]; !ok {
		return ErrNotFound
	}
	r.events[e.ID] = e
	return nil
}

func (r *inMemoryRepo) Delete(ctx context.Context, userID int64, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ev, ok := r.events[id]; !ok || ev.UserID != userID {
		return ErrNotFound
	}
	delete(r.events, id)
	return nil
}

func (r *inMemoryRepo) findRange(start, end time.Time, userID int64) []entity.Event {
	res := make([]entity.Event, 0)
	for _, e := range r.events {
		if e.UserID != userID {
			continue
		}
		if !e.Date.Before(start) && !e.Date.After(end) {
			res = append(res, e)
		}
	}
	return res
}

func (r *inMemoryRepo) FindByDay(ctx context.Context, userID int64, day time.Time) ([]entity.Event, error) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24*time.Hour - time.Nanosecond)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.findRange(start, end, userID), nil
}

func (r *inMemoryRepo) FindByWeek(ctx context.Context, userID int64, day time.Time) ([]entity.Event, error) {
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := day.AddDate(0, 0, -(weekday - 1))
	start := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.findRange(start, end, userID), nil
}

func (r *inMemoryRepo) FindByMonth(ctx context.Context, userID int64, day time.Time) ([]entity.Event, error) {
	start := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.findRange(start, end, userID), nil
}
