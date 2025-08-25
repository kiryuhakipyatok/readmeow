package models

import "github.com/google/uuid"

type Widget struct {
	Id          uuid.UUID      `json:"id"`
	Title       string         `json:"title"`
	Image       string         `json:"image"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	Tags        map[string]any `json:"tags"`
	Link        string         `json:"link"`
	Likes       uint32         `json:"likes"`
	NumOfUsers  uint32         `json:"num_of_users"`
}
