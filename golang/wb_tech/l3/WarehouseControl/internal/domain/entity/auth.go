package entity

import "time"

type Actor struct {
	Username string
	Role     Role
}

type AuthSession struct {
	Token     string
	ExpiresAt time.Time
	Actor     Actor
}
