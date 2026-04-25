package application

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/interfaces"
)

type WarehouseService struct {
	repo interfaces.WarehouseRepository
}

func NewWarehouseService(repo interfaces.WarehouseRepository) *WarehouseService {
	return &WarehouseService{repo: repo}
}

func (s *WarehouseService) CreateItem(ctx context.Context, actor entity.Actor, input entity.ItemMutation) (entity.Item, error) {
	if !actor.Role.CanManageItems() {
		return entity.Item{}, ErrForbidden
	}

	cleaned, err := sanitizeItemInput(input)
	if err != nil {
		return entity.Item{}, err
	}

	now := time.Now().UTC()
	item := entity.Item{
		ID:          uuid.NewString(),
		Name:        cleaned.Name,
		SKU:         cleaned.SKU,
		Quantity:    cleaned.Quantity,
		Location:    cleaned.Location,
		Description: cleaned.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateItem(ctx, actor, item); err != nil {
		return entity.Item{}, err
	}

	return item, nil
}

func (s *WarehouseService) GetItem(ctx context.Context, actor entity.Actor, itemID string) (entity.Item, error) {
	if !actor.Role.CanViewItems() {
		return entity.Item{}, ErrForbidden
	}

	id := strings.TrimSpace(itemID)
	if id == "" {
		return entity.Item{}, ErrInvalidInput
	}

	return s.repo.GetItem(ctx, id)
}

func (s *WarehouseService) ListItems(ctx context.Context, actor entity.Actor, filter entity.ItemFilter) ([]entity.Item, error) {
	if !actor.Role.CanViewItems() {
		return nil, ErrForbidden
	}

	filter.Query = strings.TrimSpace(filter.Query)

	return s.repo.ListItems(ctx, filter)
}

func (s *WarehouseService) UpdateItem(ctx context.Context, actor entity.Actor, itemID string, input entity.ItemMutation) (entity.Item, error) {
	if !actor.Role.CanManageItems() {
		return entity.Item{}, ErrForbidden
	}

	id := strings.TrimSpace(itemID)
	if id == "" {
		return entity.Item{}, ErrInvalidInput
	}

	cleaned, err := sanitizeItemInput(input)
	if err != nil {
		return entity.Item{}, err
	}

	currentItem, err := s.repo.GetItem(ctx, id)
	if err != nil {
		return entity.Item{}, err
	}

	currentItem.Name = cleaned.Name
	currentItem.SKU = cleaned.SKU
	currentItem.Quantity = cleaned.Quantity
	currentItem.Location = cleaned.Location
	currentItem.Description = cleaned.Description
	currentItem.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateItem(ctx, actor, currentItem); err != nil {
		return entity.Item{}, err
	}

	return currentItem, nil
}

func (s *WarehouseService) DeleteItem(ctx context.Context, actor entity.Actor, itemID string) error {
	if !actor.Role.CanDeleteItems() {
		return ErrForbidden
	}

	id := strings.TrimSpace(itemID)
	if id == "" {
		return ErrInvalidInput
	}

	return s.repo.DeleteItem(ctx, actor, id)
}

func (s *WarehouseService) ListHistory(ctx context.Context, actor entity.Actor, filter entity.HistoryFilter) ([]entity.HistoryEntry, error) {
	if !actor.Role.CanViewHistory() {
		return nil, ErrForbidden
	}
	if !filter.Action.IsValid() {
		return nil, ErrInvalidInput
	}
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return nil, ErrInvalidInput
	}

	filter.ItemID = strings.TrimSpace(filter.ItemID)
	filter.Username = strings.TrimSpace(filter.Username)

	entries, err := s.repo.ListHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	for idx := range entries {
		entries[idx].Changes = buildChanges(entries[idx])
	}

	return entries, nil
}

func sanitizeItemInput(input entity.ItemMutation) (entity.ItemMutation, error) {
	name := strings.TrimSpace(input.Name)
	sku := strings.ToUpper(strings.TrimSpace(input.SKU))
	location := strings.TrimSpace(input.Location)
	description := strings.TrimSpace(input.Description)

	switch {
	case name == "":
		return entity.ItemMutation{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	case sku == "":
		return entity.ItemMutation{}, fmt.Errorf("%w: sku is required", ErrInvalidInput)
	case location == "":
		return entity.ItemMutation{}, fmt.Errorf("%w: location is required", ErrInvalidInput)
	case input.Quantity < 0:
		return entity.ItemMutation{}, fmt.Errorf("%w: quantity must be non-negative", ErrInvalidInput)
	}

	return entity.ItemMutation{
		Name:        name,
		SKU:         sku,
		Quantity:    input.Quantity,
		Location:    location,
		Description: description,
	}, nil
}

func buildChanges(entry entity.HistoryEntry) []entity.HistoryChange {
	keys := make([]string, 0, len(entry.OldData)+len(entry.NewData))
	seen := make(map[string]struct{}, len(entry.OldData)+len(entry.NewData))

	for key := range entry.OldData {
		if shouldSkipDiffField(key) {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	for key := range entry.NewData {
		if shouldSkipDiffField(key) {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	slices.Sort(keys)

	changes := make([]entity.HistoryChange, 0, len(keys))
	for _, key := range keys {
		before := stringifyValue(entry.OldData[key])
		after := stringifyValue(entry.NewData[key])
		if before == after {
			continue
		}

		changes = append(changes, entity.HistoryChange{
			Field:  key,
			Before: before,
			After:  after,
		})
	}

	return changes
}

func shouldSkipDiffField(field string) bool {
	return field == "updated_at"
}

func stringifyValue(value any) string {
	if value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		if float64(int(typed)) == typed {
			return fmt.Sprintf("%d", int(typed))
		}
		return fmt.Sprintf("%v", typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		raw, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", typed)
		}
		return string(raw)
	}
}
