package entity

import "time"

type Event struct {
	ID           string     `json:"id"`
	UserID       int64      `json:"user_id"`
	Date         time.Time  `json:"date"`
	Title        string     `json:"event"`
	ReminderAt   *time.Time `json:"remind_at,omitempty"`
	ReminderSent bool       `json:"reminder_sent"`
	Archived     bool       `json:"archived"`
	ArchivedAt   *time.Time `json:"archived_at,omitempty"`
}
