package handlers

import (
	"readmeow/internal/delivery/dto"
	"readmeow/internal/domain/services"
	"readmeow/pkg/validator"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := th.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	if err := th.TemplateServ.Create(ctx, req.OwnerId, req.Title, req.Image, req.Description, req.Text, req.Links, req.Order, req.Widgets); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to create template: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (th *TemplateHandl) UpdateTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.UpdateTemplateRequest{}
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := th.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	if err := th.TemplateServ.Update(ctx, req.Updates, req.Id); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to update template: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (th *TemplateHandl) DeleteTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := uuid.Validate(id); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "invalid id: " + err.Error(),
		})
	}
	if err := th.TemplateServ.Delete(ctx, id); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to delete template: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (th *TemplateHandl) GetTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := uuid.Validate(id); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "invalid id: " + err.Error(),
		})
	}
	template, err := th.TemplateServ.Get(ctx, id)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get template: " + err.Error(),
		})
	}
	return c.JSON(template)
}

func (th *TemplateHandl) FetchTemplates(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.PaginationRequest{}
	if err := c.QueryParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := th.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	templates, err := th.TemplateServ.Fetch(ctx, req.Amount, req.Page)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to fetch templates: " + err.Error(),
		})
	}

	return c.JSON(templates)
}

func (th *TemplateHandl) SortTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SortTemplatesRequest{}
	if err := c.QueryParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := th.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	templates, err := th.TemplateServ.Sort(ctx, req.Amount, req.Page, strings.ToUpper(req.Destination), req.Field)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to sort templates: " + err.Error(),
		})
	}
	return c.JSON(templates)
}

func (th *TemplateHandl) SearchTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SearchRequest{}
	if err := c.QueryParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := th.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	templates, err := th.TemplateServ.Search(ctx, req.Amount, req.Page, req.Query)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to search templates: " + err.Error(),
		})
	}
	return c.JSON(templates)
}

func (th *TemplateHandl) Like(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := uuid.Validate(id); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "invalid id: " + err.Error(),
		})
	}
	cookie := c.Cookies("jwt")
	uid, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get user id: " + err.Error(),
		})
	}
	if err := th.TemplateServ.Like(ctx, id, uid); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to like template: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (th *TemplateHandl) Dislike(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := uuid.Validate(id); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "invalid id: " + err.Error(),
		})
	}
	cookie := c.Cookies("jwt")
	uid, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get user id: " + err.Error(),
		})
	}
	if err := th.TemplateServ.Dislike(ctx, id, uid); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to dislike template: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (th *TemplateHandl) FetchFavoriteTemplates(c *fiber.Ctx) error {
	ctx := c.UserContext()
	cookie := c.Cookies("jwt")
	id, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get user id: " + err.Error(),
		})
	}
	templates, err := th.TemplateServ.FetchFavorite(ctx, id)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to fetch favorite templates: " + err.Error(),
		})
	}
	return c.JSON(templates)
}
