package handlers

import (
	"encoding/json"
	"readmeow/internal/delivery/handlers/helpers"
	"readmeow/internal/domain/services"
	"readmeow/internal/domain/utils"
	"readmeow/internal/dto"
	"readmeow/pkg/validator"

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

// CreateWidget godoc
// @Summary      Create Widget
// @Description  Creating a new widget
// @Tags         Widgets
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        data formData dto.CreateWidgetRequestDoc true "Widget creation request"
// @Success      200 {object} dto.IdResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure 	 409 {object} helpers.ApiErr "Already exists"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/widgets [post]
func (wh *WidgetHandl) CreateWidget(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.CreateWidgetRequest{}
	req.Title = c.FormValue("title")
	req.Description = c.FormValue("description")
	req.Type = c.FormValue("type")
	req.Link = c.FormValue("link")
	tagsStr := c.FormValue("tags")
	tags := make(map[string]any)
	if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
		return helpers.InvalidRequest()
	}
	req.Tags = tags
	if image, _ := c.FormFile("image"); image != nil {
		req.Image = image
	}
	if errs := helpers.ValidateStruct(&req, wh.Validator); errs != nil {
		return helpers.ValidationError(errs)
	}
	id, err := wh.WidgetServ.Create(ctx, req.Title, req.Description, req.Link, req.Type, req.Tags, req.Image)
	if err != nil {
		return helpers.ToApiError(err)
	}
	idResp := dto.IdResponse{
		Id:      id,
		Message: "widget created successfully",
	}
	return c.JSON(idResp)
}

// GetWidgetById godoc
// @Summary      Get Widget
// @Description  Get widget by ID
// @Tags         Widgets
// @Produce      json
// @Security     ApiKeyAuth
// @Param        widget path string true "Widget ID"
// @Success      200 {object} models.Widget "Widget data"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/widgets/{widget} [get]
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

// SearchWidgets godoc
// @Summary      Search Widgets
// @Description  Searching widgets
// @Tags         Widgets
// @Accept       json
// @Produce      json
// @Param        body body dto.SearchWidgetRequestDoc true "Search widgets request"
// @Success      200 {array} dto.WidgetResponse "List of widgets"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/widgets [get]
func (wh *WidgetHandl) SearchWidgets(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SearchWidgetRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, wh.Validator); err != nil {
		return err
	}
	widgets, err := wh.WidgetServ.Search(ctx, req.Amount, req.Page, req.Query, req.Filter, req.Sort)
	if err != nil {
		return helpers.ToApiError(err)
	}

	return c.JSON(widgets)
}

// LikeWidget godoc
// @Summary      Like Widget
// @Description  Like widget by ID
// @Tags         Widgets
// @Produce      json
// @Security     ApiKeyAuth
// @Param        widget path string true "Widget ID"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      409 {object} helpers.ApiErr "Already exists"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/widgets/like/{widget} [patch]
func (wh *WidgetHandl) Like(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("widget")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	uid, err := utils.GetIdFromLocals(c.Locals("user"))
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := wh.WidgetServ.Like(ctx, id, uid); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// DislikeWidget godoc
// @Summary      Dislike Widget
// @Description  Dislike widget by ID
// @Tags         Widgets
// @Produce      json
// @Security     ApiKeyAuth
// @Param        widget path string true "Widget ID"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      409 {object} helpers.ApiErr "Already exists"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/widgets/dislike/{widget} [patch]
func (wh *WidgetHandl) Dislike(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("widget")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	uid, err := utils.GetIdFromLocals(c.Locals("user"))
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := wh.WidgetServ.Dislike(ctx, id, uid); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// FetchFavoriteWidgets godoc
// @Summary      Fetch Favorite Widgets
// @Description  Fetch favorite widgets of current user
// @Tags         Widgets
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body query dto.PaginationRequest true "Pagination request"
// @Success      200 {array} dto.WidgetResponse "List of favorite widgets"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/widgets/favorite [get]
func (wh *WidgetHandl) FetchFavoriteWidgets(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.PaginationRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Query{}, wh.Validator); err != nil {
		return err
	}
	id, err := utils.GetIdFromLocals(c.Locals("user"))
	if err != nil {
		return helpers.ToApiError(err)
	}
	widgets, err := wh.WidgetServ.FetchFavorite(ctx, id, req.Amount, req.Page)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(widgets)
}
