package dto

import "time"

type Note struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	Recipient      string    `json:"recipient"`
	Channel        string    `json:"channel"`
	ExpirationTime time.Time `json:"expiration_time"`
	Status         string    `json:"status"`
	Retries        int       `json:"retries"`
}
