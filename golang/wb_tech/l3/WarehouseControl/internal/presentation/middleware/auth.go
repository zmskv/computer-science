package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/interfaces"
)

const actorContextKey = "authenticated_actor"

var ErrActorMissing = errors.New("authenticated actor missing in context")

func Auth(authService interfaces.AuthService) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		header := strings.TrimSpace(c.Request.Header.Get("Authorization"))
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginext.H{"error": application.ErrUnauthorized.Error()})
			return
		}

		if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginext.H{"error": application.ErrUnauthorized.Error()})
			return
		}

		token := strings.TrimSpace(header[len("Bearer "):])
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginext.H{"error": application.ErrUnauthorized.Error()})
			return
		}

		actor, err := authService.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginext.H{"error": application.ErrUnauthorized.Error()})
			return
		}

		c.Set(actorContextKey, actor)
		c.Next()
	}
}

func ActorFromContext(c *ginext.Context) (entity.Actor, error) {
	value, ok := c.Get(actorContextKey)
	if !ok {
		return entity.Actor{}, ErrActorMissing
	}

	actor, ok := value.(entity.Actor)
	if !ok {
		return entity.Actor{}, ErrActorMissing
	}

	return actor, nil
}
