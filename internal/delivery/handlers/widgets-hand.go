package handlers

import (
	"readmeow/internal/delivery/dto"
	"readmeow/internal/domain/services"
	"readmeow/pkg/validator"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type WidgetHandl struct {
	WidgetServ services.WidgetServ
	AuthServ   services.AuthServ
	Validator  *validator.Validator
}

func NewWidgetHandl(ws services.WidgetServ, as services.AuthServ, v *validator.Validator) *WidgetHandl {
	return &WidgetHandl{
		WidgetServ: ws,
		AuthServ:   as,
		Validator:  v,
	}
}

func (wh *WidgetHandl) GetWidgetById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("widget")
	if err := ValidateId(c, id); err != nil {
		return err
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
	if err := ParseAndValidateRequest(c, &req, Query{}, wh.Validator); err != nil {
		return err
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
	if err := ParseAndValidateRequest(c, &req, Query{}, wh.Validator); err != nil {
		return err
	}
	widgets, err := wh.WidgetServ.Sort(ctx, req.Amount, req.Page, req.Field, strings.ToUpper(req.Destination))
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
	if err := ParseAndValidateRequest(c, &req, Query{}, wh.Validator); err != nil {
		return err
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
	if err := ValidateId(c, id); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := wh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get user id: " + err.Error(),
		})
	}
	if err := wh.WidgetServ.Like(ctx, id, uid); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to like widget: " + err.Error(),
		})
	}
	return SuccessResponse(c)
}

func (wh *WidgetHandl) Dislike(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("widget")
	if err := ValidateId(c, id); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := wh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get user id: " + err.Error(),
		})
	}
	if err := wh.WidgetServ.Dislike(ctx, id, uid); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to dislike widget: " + err.Error(),
		})
	}
	return SuccessResponse(c)
}

func (wh *WidgetHandl) FetchFavoriteWidgets(c *fiber.Ctx) error {
	ctx := c.UserContext()
	cookie := c.Cookies("jwt")
	id, err := wh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get user id: " + err.Error(),
		})
	}
	widgets, err := wh.WidgetServ.FetchFavorite(ctx, id)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to fetch favorite widgets: " + err.Error(),
		})
	}
	return c.JSON(widgets)
}
