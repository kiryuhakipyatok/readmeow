package models

import (
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
	Widgets        []string
	Likes          uint16
	NumOfUsers     uint16
	Order          string
	CreateTime     int64
	LastUpdateTime int64
}
