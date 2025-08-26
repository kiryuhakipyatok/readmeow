package handlers

import (
	"encoding/json"
	"readmeow/internal/delivery/handlers/helpers"
	"readmeow/internal/domain/services"
	"readmeow/internal/dto"
	"readmeow/pkg/validator"

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

// CreateTemplate godoc
// @Summary      Create Template
// @Description  Creating a new template
// @Tags         Templates
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        data formData dto.CreateTemplateRequestDoc true "Template creation request"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure 409 {object} helpers.ApiErr "Already exists"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/templates [post]
func (th *TemplateHandl) CreateTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	cookie := c.Cookies("jwt")
	req := dto.CreateTemplateRequest{}
	oid, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	req.Title = c.FormValue("title")
	req.Description = c.FormValue("description")
	form, err := c.MultipartForm()
	if err != nil {
		return helpers.InvalidRequest()
	}
	if renderOrder, ok := form.Value["render_order"]; ok {
		req.RenderOrder = renderOrder
	}
	req.Text = []string{}
	if text, ok := form.Value["text"]; ok {
		req.Text = text
	}
	req.Links = []string{}
	if links, ok := form.Value["links"]; ok {
		req.Links = links
	}
	widgetsData := c.FormValue("widgets")
	req.Widgets = []map[string]string{}
	if widgetsData != "" {
		widgets := make([]map[string]string, 0)
		if err := json.Unmarshal([]byte(widgetsData), &widgets); err != nil {
			return helpers.ToApiError(err)
		}
		req.Widgets = widgets
	}

	if image, _ := c.FormFile("image"); image != nil {
		req.Image = image
	}

	if errs := helpers.ValidateStruct(req, th.Validator); len(errs) > 0 {
		return helpers.ValidationError(errs)
	}
	if err := th.TemplateServ.Create(ctx, oid, req.Title, req.Description, req.Image, req.Links, req.RenderOrder, req.Text, req.Widgets); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// UpdateTemplate godoc
// @Summary      Update Template
// @Description  Updating a user template
// @Tags         Templates
// @Produce      json
// @Security     ApiKeyAuth
// @Param        data formData dto.UpdateTemplateRequestDoc true "Update template request"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/templates [patch]
func (th *TemplateHandl) UpdateTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	updates := make(map[string]any)
	id := c.FormValue("id")
	title := c.FormValue("title")
	if title != "" {
		updates["title"] = title
	}
	description := c.FormValue("description")
	if description != "" {
		updates["description"] = description
	}
	form, err := c.MultipartForm()
	if err != nil {
		return helpers.InvalidRequest()
	}
	if renderOrder, ok := form.Value["render_order"]; ok {
		updates["render_order"] = renderOrder
	}
	if text, ok := form.Value["text"]; ok {
		updates["text"] = text
	}
	if links, ok := form.Value["links"]; ok {
		updates["links"] = links
	}
	widgetsData := c.FormValue("widgets")
	if widgetsData != "" {
		widgets := make([]map[string]string, 0)
		if err := json.Unmarshal([]byte(widgetsData), &widgets); err != nil {
			return helpers.ToApiError(err)
		}

		updates["widgets"] = widgets
	}

	if image, _ := c.FormFile("image"); image != nil {
		updates["image"] = image
	}
	req := dto.UpdateTemplateRequest{
		Updates: updates,
		Id:      id,
	}
	if errs := helpers.ValidateStruct(req, th.Validator); len(errs) > 0 {
		return helpers.ValidationError(errs)
	}

	if err := th.TemplateServ.Update(ctx, req.Updates, req.Id); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// DeleteTemplate godoc
// @Summary      Delete Template
// @Description  Deleting a user template by its id
// @Tags         Templates
// @Produce      json
// @Security     ApiKeyAuth
// @Param        template path string true "Template ID"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/templates/{template} [delete]
func (th *TemplateHandl) DeleteTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	cookie := c.Cookies("jwt")
	uid, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	if err := th.TemplateServ.Delete(ctx, id, uid); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// GetTemplate godoc
// @Summary      Get Template
// @Description  Get template by ID
// @Tags         Templates
// @Produce      json
// @Security     ApiKeyAuth
// @Param        template path string true "Template ID"
// @Success      200 {object} dto.TemplateResponse "Template data"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/templates/{template} [get]
func (th *TemplateHandl) GetTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	template, err := th.TemplateServ.Get(ctx, id)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(template)
}

// SearchTemplates godoc
// @Summary      Search Templates
// @Description  Search templates with filters and sorting
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body body dto.SearchTemplateRequest true "Search templates request"
// @Success      200 {array} dto.TemplateResponse "List of templates"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/templates [get]
func (th *TemplateHandl) SearchTemplate(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SearchTemplateRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, th.Validator); err != nil {
		return err
	}
	templates, err := th.TemplateServ.Search(ctx, req.Amount, req.Page, req.Query, req.Filter, req.Sort)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(templates)
}

// LikeTemplate godoc
// @Summary      Like Template
// @Description  Like template by ID
// @Tags         Templates
// @Produce      json
// @Security     ApiKeyAuth
// @Param        template path string true "Template ID"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      409 {object} helpers.ApiErr "Already liked"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/templates/like/{template} [patch]
func (th *TemplateHandl) Like(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := th.TemplateServ.Like(ctx, id, uid); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// DislikeTemplate godoc
// @Summary      Dislike Template
// @Description  Dislike template by ID
// @Tags         Templates
// @Produce      json
// @Security     ApiKeyAuth
// @Param        template path string true "Template ID"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      409 {object} helpers.ApiErr "Already disliked"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/templates/dislike/{template} [patch]
func (th *TemplateHandl) Dislike(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("template")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := th.TemplateServ.Dislike(ctx, id, uid); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// FetchFavoriteTemplates godoc
// @Summary      Fetch Favorite Templates
// @Description  Fetch favorite templates of current user
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body query dto.PaginationRequest true "Pagination request"
// @Success      200 {array} dto.TemplateResponse "List of favorite templates"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/templates/favorite [get]
func (th *TemplateHandl) FetchFavoriteTemplates(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.PaginationRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Query{}, th.Validator); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	id, err := th.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	templates, err := th.TemplateServ.FetchFavorite(ctx, id, req.Amount, req.Page)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(templates)
}
