package dto

import "time"

type Comment struct {
	Id       string    `db:"id"`
	ParentID string    `db:"parent_id"`
	Text     string    `db:"text"`
	Author   string    `db:"author"`
	Date     time.Time `db:"date"`
}
