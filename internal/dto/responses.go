package dto

import (
	"time"
)

type LoginResponse struct {
	Id       string `json:"id" validate:"required,uuid"`
	Nickname string `json:"nickname" validate:"required"`
	Avatar   string `json:"avatar" validate:"required"`
}

type WidgetResponse struct {
	Id          string `json:"id" validate:"required,uuid"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Image       string `json:"image" validate:"required"`
	Likes       int32  `json:"likes" validate:"required,min=0"`
	NumOfUsers  int32  `json:"num_of_users" validate:"required,min=0"`
}

type TemplateResponse struct {
	Id             string    `json:"id" validate:"required,uuid"`
	Title          string    `json:"title" validate:"required"`
	Image          string    `json:"image" validate:"required"`
	LastUpdateTime time.Time `json:"last_update_time" validate:"required"`
	NumOfUsers     int32     `json:"num_of_users" validate:"required,min=0"`
	Likes          int32     `json:"likes" validate:"required,min=0"`
	OwnerId        string    `json:"owner_id" validate:"required,uuid"`
	OwnerAvatar    string    `json:"owner_avatar" validate:"required"`
	OwnerNickname  string    `json:"owner_nickname" validate:"required"`
}

type ReadmeResponse struct {
	Id             string    `json:"id" validate:"required,uuid"`
	Title          string    `json:"title" validate:"required"`
	Image          string    `json:"image" validate:"required"`
	LastUpdateTime time.Time `json:"last_update_time" validate:"required"`
	CreateTime     time.Time `json:"create_time" validate:"required"`
}
