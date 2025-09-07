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
	Likes       uint32 `json:"likes" validate:"required,min=0"`
	NumOfUsers  uint32 `json:"num_of_users" validate:"required,min=0"`
}

type OwnerInfo struct {
	OwnerId       string `json:"owner_id" validate:"required,uuid"`
	OwnerAvatar   string `json:"owner_avatar" validate:"required"`
	OwnerNickname string `json:"owner_nickname" validate:"required"`
}

type TemplateInfo struct {
	Id             string    `json:"id" validate:"required,uuid"`
	Title          string    `json:"title" validate:"required,min=1"`
	Description    string    `json:"description" validate:"required,min=1"`
	Image          string    `json:"image" validate:"required"`
	LastUpdateTime time.Time `json:"last_update_time" validate:"required"`
	NumOfUsers     uint32    `json:"num_of_users" validate:"required,min=0"`
	Likes          uint32    `json:"likes" validate:"required,min=0"`
	IsPublic       bool      `json:"is_public" validate:"required"`
}

type TemplateResponse struct {
	TemplateInfo
	OwnerInfo
}

type ReadmeResponse struct {
	Id             string    `json:"id" validate:"required,uuid"`
	Title          string    `json:"title" validate:"required"`
	Image          string    `json:"image" validate:"required"`
	LastUpdateTime time.Time `json:"last_update_time" validate:"required"`
	CreateTime     time.Time `json:"create_time" validate:"required"`
}

type UserResponse struct {
	Id             string         `json:"id" validate:"required,uuid"`
	Nickname       string         `json:"nickname" validate:"required,min=1"`
	Email          string         `json:"email" validate:"required,email"`
	Avatar         string         `json:"avatar" validate:"required"`
	NumOfReadmes   uint32         `json:"num_of_readmes" validate:"required,min=0"`
	NumOfTemplates uint32         `json:"num_of_templates" validate:"required,min=0"`
	Templates      []TemplateInfo `json:"templates" validate:"required,min=0"`
	TimeOfRegister time.Time      `json:"time_of_register" validate:"required"`
}

type SuccessResponse struct {
	Code    int    `json:"code" example:"200"`
	Message string `json:"message" example:"success"`
}

type IdResponse struct {
	Id      string `json:"id" example:"00000000-0000-0000-0000-000000000000" validate:"required,uuid"`
	Message string `json:"message" example:"swth created successfully"`
}
