package models

import "github.com/google/uuid"

type Widget struct {
	Id          uuid.UUID
	Title       string
	Image       string
	Description string
	Type        string
	Link        string
	Likes       uint16
	NumOfUsers  uint16
}
