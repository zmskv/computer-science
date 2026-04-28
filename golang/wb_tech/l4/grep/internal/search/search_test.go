package search

import (
	"testing"
)

func TestFilterFixedString(t *testing.T) {
	lines := []Line{
		{Number: 1, Text: "alpha"},
		{Number: 2, Text: "error: disk full"},
		{Number: 3, Text: "beta"},
	}

	got, err := Filter(lines, Options{Pattern: "error", Fixed: true})
	if err != nil {
		t.Fatalf("Filter() error = %v", err)
	}

	if len(got) != 1 || got[0].Number != 2 {
		t.Fatalf("Filter() = %+v, want line 2 only", got)
	}
}

func TestFilterIgnoreCase(t *testing.T) {
	lines := []Line{
		{Number: 1, Text: "Timeout"},
		{Number: 2, Text: "timeout"},
	}

	got, err := Filter(lines, Options{Pattern: "timeout", Fixed: true, IgnoreCase: true})
	if err != nil {
		t.Fatalf("Filter() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("Filter() matched %d lines, want 2", len(got))
	}
}

func TestFilterInvert(t *testing.T) {
	lines := []Line{
		{Number: 1, Text: "match"},
		{Number: 2, Text: "skip"},
	}

	got, err := Filter(lines, Options{Pattern: "match", Fixed: true, Invert: true})
	if err != nil {
		t.Fatalf("Filter() error = %v", err)
	}

	if len(got) != 1 || got[0].Number != 2 {
		t.Fatalf("Filter() = %+v, want line 2 only", got)
	}
}

func TestFormat(t *testing.T) {
	lines := []Line{
		{Source: "app.log", Number: 7, Text: "error"},
	}

	got := Format(lines, true, true)
	want := "app.log:7:error\n"

	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}
