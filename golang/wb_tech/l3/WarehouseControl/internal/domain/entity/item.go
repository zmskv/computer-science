package entity

import "time"

type Item struct {
	ID          string
	Name        string
	SKU         string
	Quantity    int
	Location    string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ItemMutation struct {
	Name        string
	SKU         string
	Quantity    int
	Location    string
	Description string
}

type ItemFilter struct {
	Query string
}
