package dto

import (
	"mime/multipart"
)

type VerifyRequest struct {
	Nickname string `json:"nickname" validate:"required,min=2,max=80"`
	Login    string `json:"login" validate:"required,min=2,max=80"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,min=6"`
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required,min=2,max=80"`
	Password string `json:"password" validate:"required,min=8"`
}

type PaginationRequest struct {
	Amount uint `json:"amount" validate:"required,min=1"`
	Page   uint `json:"page" validate:"required,min=1"`
}

type sortWidgetsFields struct {
	Likes      string `json:"Likes,omitempty" example:"desc/asc"`
	NumOfUsers string `json:"NumOfUsers,omitempty" example:"desc/asc"`
}

type filterWidgetsFields struct {
	Types []string `json:"Types,omitempty"`
	Tags  []string `json:"Tags,omitempty"`
}

type SearchWidgetRequest struct {
	PaginationRequest
	Query      string   `json:"query" validate:"omitempty"`
	Likes      string   `json:"likes" validate:"omitempty,oneof=asc desc"`
	NumOfUsers string   `json:"num_of_users" validate:"omitempty,oneof=asc desc"`
	Types      []string `json:"types" validate:"omitempty"`
	Tags       []string `json:"tags" validate:"omitempty"`
}

type SearchTemplateRequest struct {
	PaginationRequest
	Query          string `json:"query" validate:"omitempty"`
	Likes          string `json:"likes" validate:"omitempty,oneof=asc desc"`
	NumOfUsers     string `json:"num_of_users" validate:"omitempty,oneof=asc desc"`
	LastUpdateTime string `json:"last_update_time" validate:"omitempty,oneof=asc desc"`
	IsOfficial     *bool  `json:"is_official" validate:"omitempty"`
}

type UpdateUserRequest struct {
	Id      string         `json:"id" validate:"required,uuid"`
	Updates map[string]any `json:"updates" validate:"required,min=1,dive,keys,oneof=nickname avatar,endkeys,required"`
}

type UpdateUserRequestDoc struct {
	Id       string `json:"id" validate:"required,uuid"`
	Nickname string `json:"nickname" validate:"omitempty,min=2,max=80"`
	Avatar   string `json:"avatar" validate:"omitempty" format:"binary"`
}

type DeleteUserRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}

type ChangePasswordRequest struct {
	OldPasswrod string `json:"old_password" validate:"required,min=8"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type CreateTemplateRequest struct {
	Title       string                `json:"title" validate:"required,min=1,max=255"`
	Image       *multipart.FileHeader `json:"image" validate:"required"`
	Description string                `json:"description" validate:"required,min=1,max=1000"`
	RenderOrder []string              `json:"render_order" validate:"required,min=1"`
	Text        []string              `json:"text" validate:"omitempty"`
	Links       []string              `json:"links" validate:"omitempty,dive,url"`
	Widgets     []map[string]string   `json:"widgets" validate:"omitempty,dive,dive,keys,uuid,endkeys,required,min=1"`
	IsPublic    bool                  `json:"is_public" validate:"required"`
}

type CreateTemplateRequestDoc struct {
	Title       string              `json:"title" validate:"required,min=1,max=255"`
	Image       string              `json:"image" validate:"required" format:"binary"`
	Description string              `json:"description" validate:"required,min=1,max=1000"`
	RenderOrder []string            `json:"render_order" validate:"required,min=1"`
	Text        []string            `json:"text" validate:"omitempty"`
	Links       []string            `json:"links" validate:"omitempty,dive,url"`
	Widgets     []map[string]string `json:"widgets" validate:"omitempty,dive,dive,keys,uuid,endkeys,required,min=1"`
	IsPublic    bool                `json:"is_public" validate:"required"`
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

type GoogleOAuthRequest struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type GitHubOAuthRequest struct {
	Id     int64  `json:"id"`
	Login  string `json:"login"`
	Avatar string `json:"avatar_url"`
	Email  string `json:"email"`
}

type SendResetPasswordTokenRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}
