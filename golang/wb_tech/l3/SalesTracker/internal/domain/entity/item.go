package entity

import (
	"errors"
	"strings"
	"time"
)

type ItemType string

const (
	ItemTypeIncome  ItemType = "income"
	ItemTypeExpense ItemType = "expense"
)

type SortField string

const (
	SortFieldOccurredAt SortField = "occurred_at"
	SortFieldAmount     SortField = "amount"
	SortFieldCategory   SortField = "category"
	SortFieldType       SortField = "type"
	SortFieldCreatedAt  SortField = "created_at"
)

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

type AnalyticsGroup string

const (
	AnalyticsGroupNone     AnalyticsGroup = ""
	AnalyticsGroupDay      AnalyticsGroup = "day"
	AnalyticsGroupWeek     AnalyticsGroup = "week"
	AnalyticsGroupMonth    AnalyticsGroup = "month"
	AnalyticsGroupCategory AnalyticsGroup = "category"
)

type Item struct {
	ID          string
	Type        ItemType
	Amount      float64
	Category    string
	Description string
	OccurredAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ItemFilter struct {
	From      *time.Time
	To        *time.Time
	Type      ItemType
	Category  string
	SortBy    SortField
	SortOrder SortDirection
}

type AnalyticsFilter struct {
	From     *time.Time
	To       *time.Time
	Type     ItemType
	Category string
	GroupBy  AnalyticsGroup
}

type AnalyticsSummary struct {
	Sum          float64
	Avg          float64
	Count        int64
	Median       float64
	Percentile90 float64
}

type AnalyticsPoint struct {
	Group string
	Label string
	Sum   float64
	Avg   float64
	Count int64
}

type AnalyticsResult struct {
	Summary AnalyticsSummary
	Points  []AnalyticsPoint
}

func (i Item) Validate() error {
	validators := []func() error{
		i.validateID,
		i.validateType,
		i.validateAmount,
		i.validateCategory,
		i.validateOccurredAt,
	}

	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

func (t ItemType) IsValid() bool {
	switch t {
	case ItemTypeIncome, ItemTypeExpense:
		return true
	default:
		return false
	}
}

func (f SortField) IsValid() bool {
	switch f {
	case SortFieldOccurredAt, SortFieldAmount, SortFieldCategory, SortFieldType, SortFieldCreatedAt:
		return true
	default:
		return false
	}
}

func (d SortDirection) IsValid() bool {
	switch d {
	case SortDirectionAsc, SortDirectionDesc:
		return true
	default:
		return false
	}
}

func (g AnalyticsGroup) IsValid() bool {
	switch g {
	case AnalyticsGroupNone, AnalyticsGroupDay, AnalyticsGroupWeek, AnalyticsGroupMonth, AnalyticsGroupCategory:
		return true
	default:
		return false
	}
}

func (i Item) validateID() error {
	if strings.TrimSpace(i.ID) == "" {
		return errors.New("item id is required")
	}

	return nil
}

func (i Item) validateType() error {
	if !i.Type.IsValid() {
		return errors.New("item type is invalid")
	}

	return nil
}

func (i Item) validateAmount() error {
	if i.Amount < 0 {
		return errors.New("amount must not be negative")
	}

	return nil
}

func (i Item) validateCategory() error {
	if strings.TrimSpace(i.Category) == "" {
		return errors.New("category is required")
	}

	return nil
}

func (i Item) validateOccurredAt() error {
	if i.OccurredAt.IsZero() {
		return errors.New("occurred_at is required")
	}

	return nil
}
