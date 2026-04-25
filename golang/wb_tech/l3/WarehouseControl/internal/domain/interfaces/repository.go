package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/entity"
)

type WarehouseRepository interface {
	Close()
	CreateItem(ctx context.Context, actor entity.Actor, item entity.Item) error
	GetItem(ctx context.Context, itemID string) (entity.Item, error)
	ListItems(ctx context.Context, filter entity.ItemFilter) ([]entity.Item, error)
	UpdateItem(ctx context.Context, actor entity.Actor, item entity.Item) error
	DeleteItem(ctx context.Context, actor entity.Actor, itemID string) error
	ListHistory(ctx context.Context, filter entity.HistoryFilter) ([]entity.HistoryEntry, error)
}
