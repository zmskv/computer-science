package dto

type CreateItemRequest struct {
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	OccurredAt  string  `json:"occurred_at"`
}

type UpdateItemRequest struct {
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	OccurredAt  string  `json:"occurred_at"`
}
