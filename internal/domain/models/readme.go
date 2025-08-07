package models

import (
	"github.com/google/uuid"
)

type Readme struct {
	Id             uuid.UUID
	OwnerId        uuid.UUID
	Title          string
	Text           []string
	Links          []string
	Widgets        []string
	Order          string
	CreateTime     int64
	LastUpdateTime int64
}
