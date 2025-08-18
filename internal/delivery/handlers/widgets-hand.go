package handlers

import (
	"readmeow/internal/delivery/handlers/helpers"
	"readmeow/internal/domain/services"
	"readmeow/internal/dto"
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
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	widget, err := wh.WidgetServ.Get(ctx, id)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(widget)
}

func (wh *WidgetHandl) FetchWidgets(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.PaginationRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Query{}, wh.Validator); err != nil {
		return err
	}
	widgets, err := wh.WidgetServ.Fetch(ctx, req.Amount, req.Page)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(widgets)
}

func (wh *WidgetHandl) FetchSortedWidgets(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SortWidgetsRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Query{}, wh.Validator); err != nil {
		return err
	}
	widgets, err := wh.WidgetServ.Sort(ctx, req.Amount, req.Page, req.Field, strings.ToUpper(req.Destination))
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(widgets)
}

func (wh *WidgetHandl) SearchWidgets(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SearchRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Query{}, wh.Validator); err != nil {
		return err
	}
	widgets, err := wh.WidgetServ.Search(ctx, req.Amount, req.Page, req.Query)
	if err != nil {
		return helpers.ToApiError(err)
	}

	return c.JSON(widgets)
}

func (wh *WidgetHandl) Like(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("widget")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := wh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := wh.WidgetServ.Like(ctx, id, uid); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

func (wh *WidgetHandl) Dislike(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("widget")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := wh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := wh.WidgetServ.Dislike(ctx, id, uid); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

func (wh *WidgetHandl) FetchFavoriteWidgets(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.PaginationRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Query{}, wh.Validator); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	id, err := wh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	widgets, err := wh.WidgetServ.FetchFavorite(ctx, id, req.Amount, req.Page)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(widgets)
}
