package dto

type CommentRequest struct {
	ParentID string `json:"parent_id"`
	Text     string `json:"text"`
	Author   string `json:"author"`
}
