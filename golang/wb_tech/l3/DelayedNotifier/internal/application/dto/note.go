package dto

import (
	"time"
)

type Payload struct {
	Type string `json:"type"`
	Data Note   `json:"data"`
}

type Note struct {
	ID             string
	Title          string
	Body           string
	Recipient      string
	Channel        string
	ExpirationTime time.Time
	Status         string
	Retries        int
}
