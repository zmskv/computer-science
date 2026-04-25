package dto

type CreateItemRequest struct {
	Name        string `json:"name"`
	SKU         string `json:"sku"`
	Quantity    int    `json:"quantity"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

type UpdateItemRequest struct {
	Name        string `json:"name"`
	SKU         string `json:"sku"`
	Quantity    int    `json:"quantity"`
	Location    string `json:"location"`
	Description string `json:"description"`
}
