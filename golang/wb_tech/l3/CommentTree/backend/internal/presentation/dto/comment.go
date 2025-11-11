package dto

type CommentRequest struct {
	ParentID string `json:"parent_id"`
	Text     string `json:"text"`
	Author   string `json:"author"`
}

type GetCommentParams struct {
	ParentID string
	Search   string
	SortBy   string
	Page     int
	PageSize int
}
