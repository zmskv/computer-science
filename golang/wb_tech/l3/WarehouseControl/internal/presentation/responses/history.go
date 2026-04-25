package responses

import (
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/entity"
)

type HistoryChangeResponse struct {
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
}

type HistoryEntryResponse struct {
	ID          int64                   `json:"id"`
	ItemID      string                  `json:"item_id"`
	Action      string                  `json:"action"`
	ChangedBy   string                  `json:"changed_by"`
	ChangedRole string                  `json:"changed_role"`
	ChangedAt   string                  `json:"changed_at"`
	OldData     map[string]any          `json:"old_data"`
	NewData     map[string]any          `json:"new_data"`
	Changes     []HistoryChangeResponse `json:"changes"`
}

func HistoryEntryFromEntity(entry entity.HistoryEntry) HistoryEntryResponse {
	changes := make([]HistoryChangeResponse, 0, len(entry.Changes))
	for _, change := range entry.Changes {
		changes = append(changes, HistoryChangeResponse{
			Field:  change.Field,
			Before: change.Before,
			After:  change.After,
		})
	}

	return HistoryEntryResponse{
		ID:          entry.ID,
		ItemID:      entry.ItemID,
		Action:      string(entry.Action),
		ChangedBy:   entry.ChangedBy,
		ChangedRole: string(entry.ChangedRole),
		ChangedAt:   entry.ChangedAt.UTC().Format(time.RFC3339),
		OldData:     entry.OldData,
		NewData:     entry.NewData,
		Changes:     changes,
	}
}

func HistoryEntriesFromEntities(entries []entity.HistoryEntry) []HistoryEntryResponse {
	result := make([]HistoryEntryResponse, 0, len(entries))
	for _, entry := range entries {
		result = append(result, HistoryEntryFromEntity(entry))
	}

	return result
}
