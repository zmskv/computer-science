package entity

import "time"

type HistoryAction string

const (
	HistoryActionInsert HistoryAction = "insert"
	HistoryActionUpdate HistoryAction = "update"
	HistoryActionDelete HistoryAction = "delete"
)

func (a HistoryAction) IsValid() bool {
	switch a {
	case "", HistoryActionInsert, HistoryActionUpdate, HistoryActionDelete:
		return true
	default:
		return false
	}
}

type HistoryFilter struct {
	ItemID   string
	Username string
	Action   HistoryAction
	From     *time.Time
	To       *time.Time
}

type HistoryChange struct {
	Field  string
	Before string
	After  string
}

type HistoryEntry struct {
	ID          int64
	ItemID      string
	Action      HistoryAction
	ChangedBy   string
	ChangedRole Role
	ChangedAt   time.Time
	OldData     map[string]any
	NewData     map[string]any
	Changes     []HistoryChange
}
