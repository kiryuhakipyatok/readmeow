package models

import "github.com/google/uuid"

type Widget struct {
	Id          uuid.UUID
	Title       string
	Image       string
	Description string
	Type        string
	Tags        map[string]any
	Link        string
	Likes       int32
	NumOfUsers  int32
}
