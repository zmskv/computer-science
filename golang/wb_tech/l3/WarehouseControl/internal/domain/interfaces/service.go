package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/entity"
)

type WarehouseService interface {
	CreateItem(ctx context.Context, actor entity.Actor, input entity.ItemMutation) (entity.Item, error)
	GetItem(ctx context.Context, actor entity.Actor, itemID string) (entity.Item, error)
	ListItems(ctx context.Context, actor entity.Actor, filter entity.ItemFilter) ([]entity.Item, error)
	UpdateItem(ctx context.Context, actor entity.Actor, itemID string, input entity.ItemMutation) (entity.Item, error)
	DeleteItem(ctx context.Context, actor entity.Actor, itemID string) error
	ListHistory(ctx context.Context, actor entity.Actor, filter entity.HistoryFilter) ([]entity.HistoryEntry, error)
}

type AuthService interface {
	Login(ctx context.Context, username string, role entity.Role) (entity.AuthSession, error)
	ParseToken(token string) (entity.Actor, error)
}
