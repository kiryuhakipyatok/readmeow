package handlers

import (
	"readmeow/internal/delivery/dto"
	"readmeow/internal/domain/services"
	"readmeow/pkg/validator"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type TemplateHandl struct {
	TemplateServ services.TemplateServ
	AuthServ     services.AuthServ
	Validator    *validator.Validator
}

func NewTemplateHandl(ts services.TemplateServ, as services.AuthServ, v *validator.Validator) *TemplateHandl {
	return &TemplateHandl{
		TemplateServ: ts,
		AuthServ:     as,
		Validator:    v,
	}
}

func (th *TemplateHandl) CreateTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.CreateTemplateRequest{}
	if err := ParseAndValidateRequest(c, &req, Body{}, th.Validator); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	oid, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return ToApiError(err)
	}
	if err := th.TemplateServ.Create(ctx, oid, req.Title, req.Image, req.Description, req.Text, req.Links, req.Order, req.Widgets); err != nil {
		return ToApiError(err)
	}
	return SuccessResponse(c)
}

func (th *TemplateHandl) UpdateTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.UpdateTemplateRequest{}
	if err := ParseAndValidateRequest(c, &req, Body{}, th.Validator); err != nil {
		return err
	}
	if err := th.TemplateServ.Update(ctx, req.Updates, req.Id); err != nil {
		return ToApiError(err)
	}
	return SuccessResponse(c)
}

func (th *TemplateHandl) DeleteTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := ValidateId(c, id); err != nil {
		return err
	}
	if err := th.TemplateServ.Delete(ctx, id); err != nil {
		return ToApiError(err)
	}
	return SuccessResponse(c)
}

func (th *TemplateHandl) GetTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := ValidateId(c, id); err != nil {
		return err
	}
	template, err := th.TemplateServ.Get(ctx, id)
	if err != nil {
		return ToApiError(err)
	}
	return c.JSON(template)
}

func (th *TemplateHandl) FetchTemplates(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.PaginationRequest{}
	if err := ParseAndValidateRequest(c, &req, Query{}, th.Validator); err != nil {
		return err
	}
	templates, err := th.TemplateServ.Fetch(ctx, req.Amount, req.Page)
	if err != nil {
		return ToApiError(err)
	}
	return c.JSON(templates)
}

func (th *TemplateHandl) SortTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SortTemplatesRequest{}
	if err := ParseAndValidateRequest(c, &req, Query{}, th.Validator); err != nil {
		return err
	}
	templates, err := th.TemplateServ.Sort(ctx, req.Amount, req.Page, strings.ToUpper(req.Destination), req.Field)
	if err != nil {
		return ToApiError(err)
	}
	return c.JSON(templates)
}

func (th *TemplateHandl) SearchTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SearchRequest{}
	if err := ParseAndValidateRequest(c, &req, Query{}, th.Validator); err != nil {
		return err
	}
	templates, err := th.TemplateServ.Search(ctx, req.Amount, req.Page, req.Query)
	if err != nil {
		return ToApiError(err)
	}
	return c.JSON(templates)
}

func (th *TemplateHandl) Like(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := ValidateId(c, id); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return ToApiError(err)
	}
	if err := th.TemplateServ.Like(ctx, id, uid); err != nil {
		return ToApiError(err)
	}
	return SuccessResponse(c)
}

func (th *TemplateHandl) Dislike(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := ValidateId(c, id); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return ToApiError(err)
	}
	if err := th.TemplateServ.Dislike(ctx, id, uid); err != nil {
		return ToApiError(err)
	}
	return SuccessResponse(c)
}

func (th *TemplateHandl) FetchFavoriteTemplates(c *fiber.Ctx) error {
	ctx := c.UserContext()
	cookie := c.Cookies("jwt")
	id, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return ToApiError(err)
	}
	templates, err := th.TemplateServ.FetchFavorite(ctx, id)
	if err != nil {
		return ToApiError(err)
	}
	return c.JSON(templates)
}
