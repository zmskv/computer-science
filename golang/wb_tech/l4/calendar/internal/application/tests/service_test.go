package tests

import (
	"context"
	"testing"
	"time"

	"calendar/internal/application"
	"calendar/internal/domain/entity"
	"calendar/internal/infrastructure/repository"
)

func TestService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewInMemoryRepo()
	svc := application.NewService(repo, nil, nil, application.Config{})

	userID := int64(1)
	date := time.Date(2025, 9, 24, 0, 0, 0, 0, time.UTC)

	e, err := svc.CreateEvent(ctx, userID, date, "Meeting", nil)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}
	if e.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	events, err := svc.EventsForDay(ctx, userID, date)
	if err != nil {
		t.Fatalf("EventsForDay failed: %v", err)
	}
	if len(events) != 1 || events[0].ID != e.ID {
		t.Fatalf("expected 1 event, got %+v", events)
	}

	updatedTitle := "Updated Meeting"
	e2, err := svc.UpdateEvent(ctx, e.ID, userID, date, updatedTitle, nil)
	if err != nil {
		t.Fatalf("UpdateEvent failed: %v", err)
	}
	if e2.Title != updatedTitle {
		t.Fatalf("expected title %s, got %s", updatedTitle, e2.Title)
	}

	weekEvents, err := svc.EventsForWeek(ctx, userID, date)
	if err != nil {
		t.Fatalf("EventsForWeek failed: %v", err)
	}
	if len(weekEvents) != 1 || weekEvents[0].ID != e.ID {
		t.Fatalf("expected 1 event in week, got %+v", weekEvents)
	}

	monthEvents, err := svc.EventsForMonth(ctx, userID, date)
	if err != nil {
		t.Fatalf("EventsForMonth failed: %v", err)
	}
	if len(monthEvents) != 1 || monthEvents[0].ID != e.ID {
		t.Fatalf("expected 1 event in month, got %+v", monthEvents)
	}

	if err := svc.DeleteEvent(ctx, userID, e.ID); err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}

	eventsAfter, _ := svc.EventsForDay(ctx, userID, date)
	if len(eventsAfter) != 0 {
		t.Fatalf("expected 0 events after delete, got %+v", eventsAfter)
	}
}

func TestService_Errors(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewInMemoryRepo()
	svc := application.NewService(repo, nil, nil, application.Config{})

	userID := int64(1)
	date := time.Date(2025, 9, 24, 0, 0, 0, 0, time.UTC)

	_, err := svc.CreateEvent(ctx, userID, date, "", nil)
	if err == nil {
		t.Fatal("expected error for empty title")
	}

	err = svc.DeleteEvent(ctx, userID, "non-existent-id")
	if err == nil {
		t.Fatal("expected error for deleting non-existent event")
	}
}

func TestService_SendsReminder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewInMemoryRepo()
	notifier := &memoryNotifier{events: make(chan entity.Event, 1)}
	svc := application.NewService(repo, notifier, nil, application.Config{
		ReminderBuffer:  4,
		ArchiveInterval: time.Hour,
		ArchiveAfter:    time.Hour,
	})
	svc.Start(ctx)

	eventDate := time.Now().UTC().Add(time.Hour)
	remindAt := time.Now().UTC().Add(50 * time.Millisecond)
	created, err := svc.CreateEvent(ctx, 7, eventDate, "Doctor", &remindAt)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	select {
	case got := <-notifier.events:
		if got.ID != created.ID {
			t.Fatalf("reminder sent for event %s, want %s", got.ID, created.ID)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected reminder to be sent")
	}

	stored, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if !stored.ReminderSent {
		t.Fatal("expected reminder_sent to be true")
	}
}

func TestService_ArchivesOldEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewInMemoryRepo()
	svc := application.NewService(repo, nil, nil, application.Config{
		ReminderBuffer:  4,
		ArchiveInterval: 20 * time.Millisecond,
		ArchiveAfter:    time.Minute,
	})
	svc.Start(ctx)

	oldDate := time.Now().UTC().Add(-2 * time.Hour)
	event, err := svc.CreateEvent(ctx, 42, oldDate, "Old event", nil)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		archived, err := svc.ArchivedEvents(ctx, 42)
		if err != nil {
			t.Fatalf("ArchivedEvents failed: %v", err)
		}
		if len(archived) == 1 && archived[0].ID == event.ID {
			dayEvents, err := svc.EventsForDay(ctx, 42, oldDate)
			if err != nil {
				t.Fatalf("EventsForDay failed: %v", err)
			}
			if len(dayEvents) != 0 {
				t.Fatalf("expected archived event to disappear from active list, got %+v", dayEvents)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("expected event to be archived")
}

type memoryNotifier struct {
	events chan entity.Event
}

func (n *memoryNotifier) Send(ctx context.Context, event entity.Event) error {
	select {
	case n.events <- event:
	default:
	}
	return nil
}
