package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"calendar/internal/domain/entity"
	"calendar/internal/domain/interfaces"
)

var (
	ErrNotFound error = errors.New("event not found")
)

type inMemoryRepo struct {
	mu       sync.RWMutex
	active   map[string]entity.Event
	archived map[string]entity.Event
}

func NewInMemoryRepo() interfaces.EventRepository {
	return &inMemoryRepo{
		active:   make(map[string]entity.Event),
		archived: make(map[string]entity.Event),
	}
}

func (r *inMemoryRepo) Create(ctx context.Context, e entity.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.active[e.ID]; ok {
		return fmt.Errorf("event already exists")
	}
	if _, ok := r.archived[e.ID]; ok {
		return fmt.Errorf("event already exists")
	}
	r.active[e.ID] = e
	return nil
}

func (r *inMemoryRepo) Update(ctx context.Context, e entity.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.active[e.ID]
	if !ok {
		return ErrNotFound
	}
	e.Archived = current.Archived
	e.ArchivedAt = current.ArchivedAt
	r.active[e.ID] = e
	return nil
}

func (r *inMemoryRepo) Delete(ctx context.Context, userID int64, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ev, ok := r.active[id]; !ok || ev.UserID != userID {
		return ErrNotFound
	}
	delete(r.active, id)
	return nil
}

func (r *inMemoryRepo) GetByID(ctx context.Context, id string) (entity.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if event, ok := r.active[id]; ok {
		return event, nil
	}
	if event, ok := r.archived[id]; ok {
		return event, nil
	}
	return entity.Event{}, ErrNotFound
}

func (r *inMemoryRepo) findRange(start, end time.Time, userID int64) []entity.Event {
	res := make([]entity.Event, 0)
	for _, e := range r.active {
		if e.UserID != userID {
			continue
		}
		if !e.Date.Before(start) && !e.Date.After(end) {
			res = append(res, e)
		}
	}

	sortEvents(res)
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

func (r *inMemoryRepo) ListArchived(ctx context.Context, userID int64) ([]entity.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]entity.Event, 0)
	for _, e := range r.archived {
		if e.UserID == userID {
			res = append(res, e)
		}
	}
	sortEvents(res)
	return res, nil
}

func (r *inMemoryRepo) ArchiveBefore(ctx context.Context, before time.Time) ([]entity.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	archived := make([]entity.Event, 0)

	for id, e := range r.active {
		if e.Date.Before(before) {
			e.Archived = true
			e.ArchivedAt = &now
			r.archived[id] = e
			delete(r.active, id)
			archived = append(archived, e)
		}
	}

	sortEvents(archived)
	return archived, nil
}

func (r *inMemoryRepo) MarkReminderSent(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if event, ok := r.active[id]; ok {
		event.ReminderSent = true
		r.active[id] = event
		return nil
	}
	if event, ok := r.archived[id]; ok {
		event.ReminderSent = true
		r.archived[id] = event
		return nil
	}

	return ErrNotFound
}

func sortEvents(events []entity.Event) {
	sort.Slice(events, func(i, j int) bool {
		if events[i].Date.Equal(events[j].Date) {
			return events[i].ID < events[j].ID
		}
		return events[i].Date.Before(events[j].Date)
	})
}
