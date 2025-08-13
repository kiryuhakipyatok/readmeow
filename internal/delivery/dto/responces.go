package dto

import (
	"time"

	"github.com/google/uuid"
)

type LoginResponse struct {
	Id     string `json:"id"`
	Login  string `json:"login"`
	Avatar string `json:"avatar"`
}

type WidgetResponse struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Likes       int32  `json:"likes"`
	NumOfUsers  int32  `json:"num_of_users"`
}

type TemplateResponse struct {
	Id             string    `json:"id"`
	Title          string    `json:"title"`
	Image          string    `json:"image"`
	LastUpdateTime time.Time `json:"last_update_time"`
	NumOfUsers     int32     `json:"num_of_users"`
	Likes          int32     `json:"likes"`
	OwnerId        uuid.UUID `json:"owner_id"`
	OwnerAvatar    string    `json:"owner_avatar"`
}

type ReadmeResponse struct {
	Id             string    `json:"id"`
	Title          string    `json:"title"`
	Image          string    `json:"image"`
	LastUpdateTime time.Time `json:"last_update_time"`
	CreateTime     time.Time `json:"create_time"`
}
