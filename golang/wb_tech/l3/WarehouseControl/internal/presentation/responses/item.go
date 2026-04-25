package responses

import (
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/entity"
)

type ItemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SKU         string `json:"sku"`
	Quantity    int    `json:"quantity"`
	Location    string `json:"location"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func ItemFromEntity(item entity.Item) ItemResponse {
	return ItemResponse{
		ID:          item.ID,
		Name:        item.Name,
		SKU:         item.SKU,
		Quantity:    item.Quantity,
		Location:    item.Location,
		Description: item.Description,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func ItemsFromEntities(items []entity.Item) []ItemResponse {
	result := make([]ItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, ItemFromEntity(item))
	}

	return result
}
