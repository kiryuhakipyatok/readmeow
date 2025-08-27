package dto

import "mime/multipart"

type VerifyRequest struct {
	Nickname string `json:"nickname" validate:"required,min=1,max=80"`
	Login    string `json:"login" validate:"required,min=1,max=80"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12"`
}

type RegisterRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,min=1"`
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required,min=1,max=80"`
	Password string `json:"password" validate:"required,min=12"`
}

type PaginationRequest struct {
	Amount uint `json:"amount" validate:"required,min=1"`
	Page   uint `json:"page" validate:"required,min=1"`
}

type SearchWidgetRequest struct {
	PaginationRequest
	Query  string              `json:"query" validate:"omitempty"`
	Sort   map[string]string   `json:"sort" validate:"omitempty,dive,keys,oneof=Likes NumOfUsers,endkeys"`
	Filter map[string][]string `json:"filter" validate:"omitempty,dive,keys,oneof=Types Tags,endkeys"`
}

type SearchTemplateRequest struct {
	PaginationRequest
	Query  string            `json:"query" validate:"omitempty"`
	Sort   map[string]string `json:"sort" validate:"omitempty,dive,keys,oneof=Likes NumOfUsers LastUpdateTime,endkeys"`
	Filter map[string]bool   `json:"filter" validate:"omitempty,dive,keys,oneof=isOfficial,endkeys"`
}

type UpdateUserRequest struct {
	Id      string         `json:"id" validate:"required,uuid"`
	Updates map[string]any `json:"updates" validate:"required,min=1,dive,keys,oneof=nickname avatar,endkeys,required"`
}

type UpdateUserRequestDoc struct {
	Id       string `json:"id" validate:"required,uuid"`
	Nickname string `json:"nickname" validate:"omitempty,min=1,max=80"`
	Avatar   string `json:"avatar" validate:"omitempty" format:"binary"`
}

type DeleteUserRequest struct {
	Password string `json:"password" validate:"required, min=12"`
}

type ChangePasswordRequest struct {
	OldPasswrod string `json:"old_password" validate:"required,min=12"`
	NewPassword string `json:"new_password" validate:"required,min=12"`
}

type CreateTemplateRequest struct {
	Title       string                `json:"title" validate:"required,min=1,max=255"`
	Image       *multipart.FileHeader `json:"image" validate:"required"`
	Description string                `json:"description" validate:"required,min=1,max=1000"`
	RenderOrder []string              `json:"render_order" validate:"required,min=1"`
	Text        []string              `json:"text" validate:"omitempty"`
	Links       []string              `json:"links" validate:"omitempty,dive,url"`
	Widgets     []map[string]string   `json:"widgets" validate:"omitempty,dive,dive,keys,uuid,endkeys,required,min=1"`
	IsPublic    bool                  `json:"is_public" validate:"omitempty,oneof=false"`
}

type CreateTemplateRequestDoc struct {
	Title       string              `json:"title" validate:"required,min=1,max=255"`
	Image       string              `json:"image" validate:"required" format:"binary"`
	Description string              `json:"description" validate:"required,min=1,max=1000"`
	RenderOrder []string            `json:"render_order" validate:"required,min=1"`
	Text        []string            `json:"text" validate:"omitempty"`
	Links       []string            `json:"links" validate:"omitempty,dive,url"`
	Widgets     []map[string]string `json:"widgets" validate:"omitempty,dive,dive,keys,uuid,endkeys,required,min=1"`
	IsPublic    bool                `json:"is_public" validate:"omitempty"`
}

type UpdateTemplateRequest struct {
	Id      string         `json:"id" validate:"required,uuid"`
	Updates map[string]any `json:"updates" validate:"required,min=1,dive,keys,oneof=title image is_public description text links widgets render_order,endkeys,required"`
}

type UpdateTemplateRequestDoc struct {
	Id          string              `json:"id" validate:"required,uuid"`
	Title       string              `json:"title" validate:"omitempty,min=1,max=255"`
	Description string              `json:"description" validate:"omitempty,min=1,max=1000"`
	Text        string              `json:"text" validate:"omitempty"`
	Links       []string            `json:"links" validate:"omitempty,dive,url"`
	Widgets     []map[string]string `json:"widgets" validate:"omitempty,dive,dive,keys,uuid,endkeys,required,min=1"`
	RenderOrder []string            `json:"render_order" validate:"omitempty"`
	Image       string              `json:"image" validate:"required" format:"binary"`
	IsPublic    bool                `json:"is_public" validate:"omitempty"`
}

type CreateReadmeRequest struct {
	TemplateId  string                `json:"template_id" validate:"omitempty,uuid"`
	Image       *multipart.FileHeader `json:"image" validate:"required"`
	Title       string                `json:"title" validate:"required,min=1,max=80"`
	RenderOrder []string              `json:"render_order" validate:"required,min=1"`
	Text        []string              `json:"text" validate:"omitempty"`
	Links       []string              `json:"links" validate:"omitempty,dive,url"`
	Widgets     []map[string]string   `json:"widgets" validate:"omitempty,dive,dive,keys,uuid,endkeys,required,min=1"`
}

type CreateReadmeRequestDoc struct {
	TemplateId  string              `json:"template_id" validate:"omitempty,uuid"`
	Image       string              `json:"image" validate:"required" format:"binary"`
	Title       string              `json:"title" validate:"required,min=1,max=80"`
	RenderOrder []string            `json:"render_order" validate:"required,min=1"`
	Text        []string            `json:"text" validate:"omitempty"`
	Links       []string            `json:"links" validate:"omitempty,dive,url"`
	Widgets     []map[string]string `json:"widgets" validate:"omitempty,dive,dive,keys,uuid,endkeys,required,min=1"`
}

type UpdateReadmeRequest struct {
	Id      string         `json:"id" validate:"required,uuid"`
	Updates map[string]any `json:"updates" validate:"required,min=1,dive,keys,oneof=title image text links widgets render_order,endkeys,required"`
}

type UpdateReadmeRequestDoc struct {
	Id          string              `json:"id" validate:"required,uuid"`
	Image       string              `json:"image" validate:"required" format:"binary"`
	Title       string              `json:"title" validate:"required,min=1,max=80"`
	RenderOrder []string            `json:"render_order" validate:"required,min=1"`
	Text        []string            `json:"text" validate:"omitempty"`
	Links       []string            `json:"links" validate:"omitempty,dive,url"`
	Widgets     []map[string]string `json:"widgets" validate:"omitempty,dive,dive,keys,uuid,endkeys,required,min=1"`
}

type SendNewCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
}
