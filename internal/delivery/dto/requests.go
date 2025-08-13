package dto

type RegisterRequest struct {
	Login    string `json:"login" validate:"required,max=80"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12"`
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required,max=80"`
	Password string `json:"password" validate:"required"`
}

type PaginationRequest struct {
	Amount uint `json:"amount" validate:"required,min=1"`
	Page   uint `json:"page" validate:"required,min=1"`
}

type SortWidgetsRequest struct {
	PaginationRequest
	Field       string `json:"field" validate:"required,oneof=likes num_of_users"`
	Destination string `json:"destination" validate:"oneof=desc asc"`
}

type SearchRequest struct {
	PaginationRequest
	Query string `json:"query" validate:"required"`
}

type UpdateUserRequest struct {
	Id      string         `json:"id" validate:"required,uuid"`
	Updates map[string]any `json:"updates" validate:"required,dive,keys,oneof=login email avatar,endkeys,required"`
}

type DeleteUserRequest struct {
	Id       string `json:"id" validate:"required,uuid"`
	Password string `json:"password" validate:"required"`
}

type ChangePasswordRequest struct {
	Id          string `json:"id" validate:"required,uuid"`
	OldPasswrod string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type CreateTemplateRequest struct {
	OwnerId     string              `json:"owner_id" validate:"required,uuid"`
	Title       string              `json:"title" validate:"required,max=50"`
	Image       string              `json:"image" validate:"required"`
	Description string              `json:"description" validate:"required,min=1,max=1000"`
	Order       []string            `json:"order" validate:"required,dive,string"`
	Text        []string            `json:"text" validate:"dive,uuid"`
	Links       []string            `json:"links" validate:"dive,uuid"`
	Widgets     []map[string]string `json:"widgets" validate:"dive,keys,uuid,endkeys,string,required"`
}

type UpdateTemplateRequest struct {
	Id      string         `json:"id" validate:"required,uuid"`
	Updates map[string]any `json:"updates" validate:"required,dive,keys,oneof=title text links widgets order,endkeys,required"`
}

type SortTemplatesRequest struct {
	PaginationRequest
	Field       string `json:"field" validate:"required,oneof=likes num_of_users create_time"`
	Destination string `json:"destination" validate:"oneof=DESC ASC"`
}

type CreateReadmeRequest struct {
	TemplateId string              `json:"template_id" validate:"uuid"`
	OwnerId    string              `json:"owner_id" validate:"required,uuid"`
	Image      string              `json:"image" validate:"required"`
	Title      string              `json:"title" validate:"required,max=80"`
	Order      []string            `json:"order" validate:"required,dive,string"`
	Text       []string            `json:"text" validate:"dive,uuid"`
	Links      []string            `json:"links" validate:"dive,uuid"`
	Widgets    []map[string]string `json:"widgets" validate:"dive,keys,uuid,endkeys,string,required"`
}

type DeleteReadmeRequest struct {
	Id      string `json:"id" validate:"required,uuid"`
	OwnerId string `json:"owner_id" validate:"required,uuid"`
}

type UpdateReadmeRequest struct {
	Id      string         `json:"id" validate:"required,uuid"`
	Updates map[string]any `json:"updates" validate:"required,dive,keys,oneof=title text links widgets order,endkeys,required"`
}

type FetchReadmeRequest struct {
	PaginationRequest
	UserId string `json:"user_id" validate:"required,uuid"`
}
