package models

import (
	"time"

	"github.com/google/uuid"
)

type Template struct {
	Id             uuid.UUID
	OwnerId        uuid.UUID
	Title          string
	Image          string
	Description    string
	Text           []string
	Links          []string
	Widgets        map[string]string
	Likes          uint16
	NumOfUsers     uint16
	Order          []string
	CreateTime     time.Time
	LastUpdateTime time.Time
}
