package entity

import "time"

type Comment struct {
	Id       string
	ParentID string
	Text     string
	Author   string
	Date     time.Time
}
