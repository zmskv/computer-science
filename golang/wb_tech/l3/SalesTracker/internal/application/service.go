package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/domain/interfaces"
	"go.uber.org/zap"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrItemNotFound = errors.New("item not found")
)

type salesService struct {
	repo   interfaces.SalesRepository
	logger *zap.Logger
}

func NewSalesService(repo interfaces.SalesRepository, logger *zap.Logger) interfaces.SalesService {
	return &salesService{
		repo:   repo,
		logger: logger,
	}
}

func (s *salesService) CreateItem(
	ctx context.Context,
	itemType entity.ItemType,
	amount float64,
	category string,
	description string,
	occurredAt time.Time,
) (entity.Item, error) {
	now := time.Now().UTC()
	item := entity.Item{
		ID:          uuid.NewString(),
		Type:        itemType,
		Amount:      amount,
		Category:    strings.TrimSpace(category),
		Description: strings.TrimSpace(description),
		OccurredAt:  occurredAt.UTC(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := validateItem(item); err != nil {
		return entity.Item{}, err
	}

	if err := s.repo.CreateItem(ctx, item); err != nil {
		return entity.Item{}, err
	}

	return item, nil
}

func (s *salesService) GetItem(ctx context.Context, itemID string) (entity.Item, error) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return entity.Item{}, fmt.Errorf("%w: item id is required", ErrInvalidInput)
	}

	return s.repo.GetItem(ctx, itemID)
}

func (s *salesService) ListItems(ctx context.Context, filter entity.ItemFilter) ([]entity.Item, error) {
	normalized, err := normalizeItemFilter(filter)
	if err != nil {
		return nil, err
	}

	return s.repo.ListItems(ctx, normalized)
}

func (s *salesService) UpdateItem(
	ctx context.Context,
	itemID string,
	itemType entity.ItemType,
	amount float64,
	category string,
	description string,
	occurredAt time.Time,
) (entity.Item, error) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return entity.Item{}, fmt.Errorf("%w: item id is required", ErrInvalidInput)
	}

	current, err := s.repo.GetItem(ctx, itemID)
	if err != nil {
		return entity.Item{}, err
	}

	current.Type = itemType
	current.Amount = amount
	current.Category = strings.TrimSpace(category)
	current.Description = strings.TrimSpace(description)
	current.OccurredAt = occurredAt.UTC()
	current.UpdatedAt = time.Now().UTC()

	if err := validateItem(current); err != nil {
		return entity.Item{}, err
	}

	if err := s.repo.UpdateItem(ctx, current); err != nil {
		return entity.Item{}, err
	}

	return current, nil
}

func (s *salesService) DeleteItem(ctx context.Context, itemID string) error {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return fmt.Errorf("%w: item id is required", ErrInvalidInput)
	}

	return s.repo.DeleteItem(ctx, itemID)
}

func (s *salesService) GetAnalytics(ctx context.Context, filter entity.AnalyticsFilter) (entity.AnalyticsResult, error) {
	normalized, err := normalizeAnalyticsFilter(filter)
	if err != nil {
		return entity.AnalyticsResult{}, err
	}

	return s.repo.GetAnalytics(ctx, normalized)
}

func validateItem(item entity.Item) error {
	if err := item.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	return nil
}

func normalizeItemFilter(filter entity.ItemFilter) (entity.ItemFilter, error) {
	filter.Category = strings.TrimSpace(filter.Category)

	if err := validateDateRange(filter.From, filter.To); err != nil {
		return entity.ItemFilter{}, err
	}

	if filter.Type != "" && !filter.Type.IsValid() {
		return entity.ItemFilter{}, fmt.Errorf("%w: item type is invalid", ErrInvalidInput)
	}

	if filter.SortBy == "" {
		filter.SortBy = entity.SortFieldOccurredAt
	}
	if !filter.SortBy.IsValid() {
		return entity.ItemFilter{}, fmt.Errorf("%w: sort_by is invalid", ErrInvalidInput)
	}

	if filter.SortOrder == "" {
		filter.SortOrder = entity.SortDirectionDesc
	}
	if !filter.SortOrder.IsValid() {
		return entity.ItemFilter{}, fmt.Errorf("%w: sort_order is invalid", ErrInvalidInput)
	}

	return filter, nil
}

func normalizeAnalyticsFilter(filter entity.AnalyticsFilter) (entity.AnalyticsFilter, error) {
	filter.Category = strings.TrimSpace(filter.Category)

	if err := validateDateRange(filter.From, filter.To); err != nil {
		return entity.AnalyticsFilter{}, err
	}

	if filter.Type != "" && !filter.Type.IsValid() {
		return entity.AnalyticsFilter{}, fmt.Errorf("%w: item type is invalid", ErrInvalidInput)
	}

	if filter.GroupBy == entity.AnalyticsGroupNone {
		filter.GroupBy = entity.AnalyticsGroupDay
	}
	if !filter.GroupBy.IsValid() {
		return entity.AnalyticsFilter{}, fmt.Errorf("%w: group_by is invalid", ErrInvalidInput)
	}

	return filter, nil
}

func validateDateRange(from, to *time.Time) error {
	if from == nil || to == nil {
		return nil
	}

	if from.After(*to) {
		return fmt.Errorf("%w: from must be before to", ErrInvalidInput)
	}

	return nil
}
