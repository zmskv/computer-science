package responses

import (
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/domain/entity"
)

type Item struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AnalyticsSummary struct {
	Sum          float64 `json:"sum"`
	Avg          float64 `json:"avg"`
	Count        int64   `json:"count"`
	Median       float64 `json:"median"`
	Percentile90 float64 `json:"percentile_90"`
}

type AnalyticsPoint struct {
	Group string  `json:"group"`
	Label string  `json:"label"`
	Sum   float64 `json:"sum"`
	Avg   float64 `json:"avg"`
	Count int64   `json:"count"`
}

type Analytics struct {
	Summary AnalyticsSummary `json:"summary"`
	Points  []AnalyticsPoint `json:"points"`
}

func ItemFromEntity(item entity.Item) Item {
	return Item{
		ID:          item.ID,
		Type:        string(item.Type),
		Amount:      item.Amount,
		Category:    item.Category,
		Description: item.Description,
		OccurredAt:  item.OccurredAt,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func ItemsFromEntities(items []entity.Item) []Item {
	response := make([]Item, 0, len(items))
	for _, item := range items {
		response = append(response, ItemFromEntity(item))
	}

	return response
}

func AnalyticsFromEntity(result entity.AnalyticsResult) Analytics {
	points := make([]AnalyticsPoint, 0, len(result.Points))
	for _, point := range result.Points {
		points = append(points, AnalyticsPoint{
			Group: point.Group,
			Label: point.Label,
			Sum:   point.Sum,
			Avg:   point.Avg,
			Count: point.Count,
		})
	}

	return Analytics{
		Summary: AnalyticsSummary{
			Sum:          result.Summary.Sum,
			Avg:          result.Summary.Avg,
			Count:        result.Summary.Count,
			Median:       result.Summary.Median,
			Percentile90: result.Summary.Percentile90,
		},
		Points: points,
	}
}
