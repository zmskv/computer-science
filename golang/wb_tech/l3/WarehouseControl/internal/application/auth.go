package application

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/entity"
)

type AuthService struct {
	secret []byte
	ttl    time.Duration
}

type tokenClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(secret string, ttl time.Duration) *AuthService {
	return &AuthService{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (s *AuthService) Login(_ context.Context, username string, role entity.Role) (entity.AuthSession, error) {
	name := strings.TrimSpace(username)
	if name == "" {
		return entity.AuthSession{}, ErrInvalidInput
	}
	if !role.IsValid() {
		return entity.AuthSession{}, ErrInvalidInput
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.ttl)

	claims := tokenClaims{
		Username: name,
		Role:     string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   name,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return entity.AuthSession{}, err
	}

	return entity.AuthSession{
		Token:     signedToken,
		ExpiresAt: expiresAt,
		Actor: entity.Actor{
			Username: name,
			Role:     role,
		},
	}, nil
}

func (s *AuthService) ParseToken(token string) (entity.Actor, error) {
	parsedClaims := &tokenClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, parsedClaims, func(parsedToken *jwt.Token) (any, error) {
		if _, ok := parsedToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}

		return s.secret, nil
	})
	if err != nil {
		return entity.Actor{}, ErrUnauthorized
	}
	if !parsedToken.Valid {
		return entity.Actor{}, ErrUnauthorized
	}

	role := entity.Role(parsedClaims.Role)
	if strings.TrimSpace(parsedClaims.Username) == "" || !role.IsValid() {
		return entity.Actor{}, ErrUnauthorized
	}

	return entity.Actor{
		Username: parsedClaims.Username,
		Role:     role,
	}, nil
}
