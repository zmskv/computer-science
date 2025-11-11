package dto

import "time"

type GetCommentParams struct {
	ParentID string
	Search   string
	SortBy   string
	Page     int
	PageSize int
}

type Comment struct {
	Id       string    `db:"id"`
	ParentID string    `db:"parent_id"`
	Text     string    `db:"text"`
	Author   string    `db:"author"`
	Date     time.Time `db:"date"`
}
