package models

import (
	"time"

	"github.com/google/uuid"
)

type Readme struct {
	Id             uuid.UUID
	OwnerId        uuid.UUID
	TemplateId     uuid.UUID
	Title          string
	Image          string
	Text           []string
	Links          []string
	Widgets        []string
	Order          string
	CreateTime     time.Time
	LastUpdateTime time.Time
}
