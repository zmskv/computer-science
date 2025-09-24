package dto

import "time"

type Event struct {
	ID     string    `json:"id"`
	UserID int64     `json:"user_id"`
	Date   time.Time `json:"date"`
	Title  string    `json:"event"`
}

type CreateEventReq struct {
	UserID int64  `json:"user_id" form:"user_id"`
	Date   string `json:"date" form:"date"`
	Title  string `json:"event" form:"event"`
}

type UpdateEventReq struct {
	ID     string `json:"id" form:"id"`
	UserID int64  `json:"user_id" form:"user_id"`
	Date   string `json:"date" form:"date"`
	Title  string `json:"event" form:"event"`
}

type DeleteEventReq struct {
	ID     string `json:"id" form:"id"`
	UserID int64  `json:"user_id" form:"user_id"`
}

type QueryReq struct {
	UserID int64  `form:"user_id" binding:"required"`
	Date   string `form:"date" binding:"required"`
}
