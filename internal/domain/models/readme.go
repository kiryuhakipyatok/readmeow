package models

import (
	"time"

	"github.com/google/uuid"
)

type Readme struct {
	Id             uuid.UUID           `json:"id"`
	OwnerId        uuid.UUID           `json:"owner_id"`
	TemplateId     uuid.UUID           `json:"template_id"`
	Title          string              `json:"title"`
	Image          string              `json:"image"`
	Text           []string            `json:"text"`
	Links          []string            `json:"links"`
	Widgets        []map[string]string `json:"widgets"`
	Order          []string            `json:"order"`
	CreateTime     time.Time           `json:"create_time"`
	LastUpdateTime time.Time           `json:"last_update_time"`
}
