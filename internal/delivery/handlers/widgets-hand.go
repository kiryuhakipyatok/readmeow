package handlers

import (
	"readmeow/internal/delivery/dto"
	"readmeow/internal/domain/services"
	"readmeow/pkg/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WidgetHandl struct {
	WidgetServ services.WidgetServ
	Validator  *validator.Validator
}

func NewWidgetHandl(ws services.WidgetServ, v *validator.Validator) *WidgetHandl {
	return &WidgetHandl{
		WidgetServ: ws,
		Validator:  v,
	}
}

func (wh *WidgetHandl) GetWidgetById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("widget")
	if err := uuid.Validate(id); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "invalid id: " + err.Error(),
		})
	}
	widget, err := wh.WidgetServ.Get(ctx, id)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get widget: " + err.Error(),
		})
	}
	return c.JSON(widget)
}

func (wh *WidgetHandl) FetchWidgets(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.PaginationRequest{}
	if err := c.QueryParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := wh.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	widgets, err := wh.WidgetServ.Fetch(ctx, req.Amount, req.Page)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to fetch widgets: " + err.Error(),
		})
	}
	return c.JSON(widgets)
}

func (wh *WidgetHandl) FetchSortedWidgets(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SortWidgetsRequest{}
	if err := c.QueryParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := wh.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	widgets, err := wh.WidgetServ.Sort(ctx, req.Amount, req.Page, req.Field, req.Destination)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to fetch sorted widgets: " + err.Error(),
		})
	}
	return c.JSON(widgets)
}

func (wh *WidgetHandl) SearchWidgets(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SearchRequest{}
	if err := c.QueryParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := wh.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	widgets, err := wh.WidgetServ.Search(ctx, req.Amount, req.Page, req.Query)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to fetch searched widgets: " + err.Error(),
		})
	}
	return c.JSON(widgets)
}

func (wh *WidgetHandl) Like(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("widget")
	if err := uuid.Validate(id); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "invalid id: " + err.Error(),
		})
	}
	if err := wh.WidgetServ.Like(ctx, id); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to like widget: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (wh *WidgetHandl) Dislike(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("widget")
	if err := uuid.Validate(id); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "invalid id: " + err.Error(),
		})
	}
	if err := wh.WidgetServ.Dislike(ctx, id); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to dislike widget: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
