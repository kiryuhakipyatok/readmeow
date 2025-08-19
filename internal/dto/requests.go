package dto

type VerifyRequest struct {
	Nickname string `json:"nickname" validate:"required,max=80"`
	Login    string `json:"login" validate:"required,max=80"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12"`
}

type RegisterRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required"`
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required,max=80"`
	Password string `json:"password" validate:"required"`
}

type PaginationRequest struct {
	Amount uint `json:"amount" validate:"required,min=1"`
	Page   uint `json:"page" validate:"required,min=1"`
}

type SearchWidgetRequest struct {
	PaginationRequest
	Query  string              `json:"query"`
	Sort   map[string]string   `json:"sort" validate:"dive,keys,oneof=Likes NumOfUsers,endkeys"`
	Filter map[string][]string `json:"filter" validate:"dive,keys,oneof=Types Tags,endkeys"`
}

type SearchTemplateRequest struct {
	PaginationRequest
	Query  string            `json:"query"`
	Sort   map[string]string `json:"sort" validate:"dive,keys,oneof=Likes NumOfUsers LastUpdateTime,endkeys"`
	Filter map[string]bool   `json:"filter" validate:"dive,keys,oneof=isOfficial,endkeys"`
}

type UpdateUserRequest struct {
	Id      string         `json:"id" validate:"required,uuid"`
	Updates map[string]any `json:"updates" validate:"required,dive,keys,oneof=login email avatar,endkeys,required"`
}

type DeleteUserRequest struct {
	Password string `json:"password" validate:"required"`
}

type ChangePasswordRequest struct {
	OldPasswrod string `json:"old_password" validate:"required,min=12"`
	NewPassword string `json:"new_password" validate:"required,min=12"`
}

type CreateTemplateRequest struct {
	Title       string              `json:"title" validate:"required,max=50"`
	Image       string              `json:"image" validate:"required"`
	Description string              `json:"description" validate:"required,min=1,max=1000"`
	Order       []string            `json:"order" validate:"required"`
	Text        []string            `json:"text"`
	Links       []string            `json:"links"`
	Widgets     []map[string]string `json:"widgets" validate:"dive,dive,keys,uuid,endkeys"`
}

type UpdateTemplateRequest struct {
	Id      string         `json:"id" validate:"required,uuid"`
	Updates map[string]any `json:"updates" validate:"required,dive,keys,oneof=title text links widgets order,endkeys,required"`
}

type CreateReadmeRequest struct {
	TemplateId string              `json:"template_id" validate:"uuid"`
	Image      string              `json:"image" validate:"required"`
	Title      string              `json:"title" validate:"required,max=80"`
	Order      []string            `json:"order" validate:"required"`
	Text       []string            `json:"text"`
	Links      []string            `json:"links"`
	Widgets    []map[string]string `json:"widgets" validate:"dive,dive,keys,uuid,endkeys,required"`
}

type UpdateReadmeRequest struct {
	Id      string         `json:"id" validate:"required,uuid"`
	Updates map[string]any `json:"updates" validate:"required,dive,keys,oneof=title text links widgets render_order,endkeys,required"`
}

type SendNewCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
}
