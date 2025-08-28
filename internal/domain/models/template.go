package models

import (
	"time"

	"github.com/google/uuid"
)

type Template struct {
	Id             uuid.UUID           `json:"id"`
	OwnerId        uuid.UUID           `json:"owner_id"`
	Title          string              `json:"title"`
	Image          string              `json:"image"`
	Description    string              `json:"description"`
	Text           []string            `json:"text"`
	Links          []string            `json:"links"`
	Widgets        []map[string]string `json:"widgets"`
	Likes          uint32              `json:"likes"`
	NumOfUsers     uint32              `json:"num_of_users"`
	RenderOrder    []string            `json:"render_order"`
	CreateTime     time.Time           `json:"create_time"`
	LastUpdateTime time.Time           `json:"last_update_time"`
	IsPublic       bool                `json:"is_public"`
}

type TemplateWithOwner struct {
	Template
	OwnerAvatar   string `json:"owner_avatar"`
	OwnerNickname string `json:"owner_nickname"`
}
