package tests

import (
	"context"
	"testing"
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/infrastructure/repository"
)

func TestService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewInMemoryRepo()
	svc := application.NewService(repo)

	userID := int64(1)
	date := time.Date(2025, 9, 24, 0, 0, 0, 0, time.UTC)

	e, err := svc.CreateEvent(ctx, userID, date, "Meeting")
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
	e2, err := svc.UpdateEvent(ctx, e.ID, userID, date, updatedTitle)
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
	svc := application.NewService(repo)

	userID := int64(1)
	date := time.Date(2025, 9, 24, 0, 0, 0, 0, time.UTC)

	_, err := svc.CreateEvent(ctx, userID, date, "")
	if err == nil {
		t.Fatal("expected error for empty title")
	}

	err = svc.DeleteEvent(ctx, userID, "non-existent-id")
	if err == nil {
		t.Fatal("expected error for deleting non-existent event")
	}
}
