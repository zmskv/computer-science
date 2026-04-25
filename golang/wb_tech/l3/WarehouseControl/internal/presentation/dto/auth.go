package dto

type LoginRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}
