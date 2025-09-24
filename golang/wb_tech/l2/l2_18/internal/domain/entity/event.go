package entity

import "time"

type Event struct {
	ID     string
	UserID int64
	Date   time.Time
	Title  string
}
