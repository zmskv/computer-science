package responses

import (
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/entity"
)

type UserResponse struct {
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type AuthSessionResponse struct {
	Token     string       `json:"token"`
	ExpiresAt string       `json:"expires_at"`
	User      UserResponse `json:"user"`
}

func AuthSessionFromEntity(session entity.AuthSession) AuthSessionResponse {
	return AuthSessionResponse{
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt.UTC().Format(time.RFC3339),
		User:      UserFromActor(session.Actor),
	}
}

func UserFromActor(actor entity.Actor) UserResponse {
	return UserResponse{
		Username:    actor.Username,
		Role:        string(actor.Role),
		Permissions: actor.Role.Permissions(),
	}
}
