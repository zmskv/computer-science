package dto

type Event struct {
	ID           string `json:"id"`
	UserID       int64  `json:"user_id"`
	Date         string `json:"date"`
	Title        string `json:"event"`
	RemindAt     string `json:"remind_at,omitempty"`
	ReminderSent bool   `json:"reminder_sent"`
	Archived     bool   `json:"archived"`
	ArchivedAt   string `json:"archived_at,omitempty"`
}

type CreateEventReq struct {
	UserID   int64  `json:"user_id" form:"user_id"`
	Date     string `json:"date" form:"date"`
	Title    string `json:"event" form:"event"`
	RemindAt string `json:"remind_at" form:"remind_at"`
}

type UpdateEventReq struct {
	ID       string `json:"id" form:"id"`
	UserID   int64  `json:"user_id" form:"user_id"`
	Date     string `json:"date" form:"date"`
	Title    string `json:"event" form:"event"`
	RemindAt string `json:"remind_at" form:"remind_at"`
}

type DeleteEventReq struct {
	ID     string `json:"id" form:"id"`
	UserID int64  `json:"user_id" form:"user_id"`
}

type QueryReq struct {
	UserID int64  `form:"user_id" binding:"required"`
	Date   string `form:"date" binding:"required"`
}

type ArchivedQueryReq struct {
	UserID int64 `form:"user_id" binding:"required"`
}
