package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/domain/entity"
)

type SalesRepository interface {
	CreateItem(ctx context.Context, item entity.Item) error
	GetItem(ctx context.Context, itemID string) (entity.Item, error)
	ListItems(ctx context.Context, filter entity.ItemFilter) ([]entity.Item, error)
	UpdateItem(ctx context.Context, item entity.Item) error
	DeleteItem(ctx context.Context, itemID string) error
	GetAnalytics(ctx context.Context, filter entity.AnalyticsFilter) (entity.AnalyticsResult, error)
	Close()
}
