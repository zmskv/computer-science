package interfaces

import (
	"context"
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/domain/entity"
)

type SalesService interface {
	CreateItem(
		ctx context.Context,
		itemType entity.ItemType,
		amount float64,
		category string,
		description string,
		occurredAt time.Time,
	) (entity.Item, error)
	GetItem(ctx context.Context, itemID string) (entity.Item, error)
	ListItems(ctx context.Context, filter entity.ItemFilter) ([]entity.Item, error)
	UpdateItem(
		ctx context.Context,
		itemID string,
		itemType entity.ItemType,
		amount float64,
		category string,
		description string,
		occurredAt time.Time,
	) (entity.Item, error)
	DeleteItem(ctx context.Context, itemID string) error
	GetAnalytics(ctx context.Context, filter entity.AnalyticsFilter) (entity.AnalyticsResult, error)
}
