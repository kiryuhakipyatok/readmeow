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

type ReadmeHandl struct {
	ReadmeServ services.ReadmeServ
	AuthServ   services.AuthServ
	Validator  *validator.Validator
}

func NewReadmeHandl(rs services.ReadmeServ, as services.AuthServ, v *validator.Validator) *ReadmeHandl {
	return &ReadmeHandl{
		ReadmeServ: rs,
		AuthServ:   as,
		Validator:  v,
	}
}

// CreateReadme godoc
// @Summary      Create Readme
// @Description  Creating a new readme
// @Tags         Readmes
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        data formData dto.CreateReadmeRequestDoc true "Readme creation request"
// @Success      200 {object} dto.IdResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure 409 {object} helpers.ApiErr "Already exists"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/readmes [post]
func (rh *ReadmeHandl) CreateReadme(c *fiber.Ctx) error {
	ctx := c.UserContext()

	req := dto.CreateReadmeRequest{}
	uid, err := utils.GetIdFromLocals(c.Locals("user"))
	if err != nil {
		return helpers.ToApiError(err)
	}

	req.TemplateId = c.FormValue("template_id")

	req.Title = c.FormValue("title")
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
			return helpers.InvalidRequest()
		}
		req.Widgets = widgets
	}
	if image, _ := c.FormFile("image"); image != nil {
		req.Image = image
	}

	if errs := helpers.ValidateStruct(req, rh.Validator); len(errs) > 0 {
		return helpers.ValidationError(errs)
	}

	id, err := rh.ReadmeServ.Create(ctx, req.TemplateId, uid, req.Title, req.Image, req.Text, req.Links, req.RenderOrder, req.Widgets)
	if err != nil {
		return helpers.ToApiError(err)
	}
	idResp := dto.IdResponse{
		Id:      id,
		Message: "readme created successfully",
	}
	return c.JSON(idResp)
}

// DeleteReadme godoc
// @Summary      Delete Readme
// @Description  Deleting a user readme by its id
// @Tags         Readmes
// @Produce      json
// @Security     ApiKeyAuth
// @Param        readme path string true "Readme ID"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/readmes/{readme} [delete]
func (rh *ReadmeHandl) DeleteReadme(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("readme")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	uid, err := utils.GetIdFromLocals(c.Locals("user"))
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := rh.ReadmeServ.Delete(ctx, id, uid); err != nil {
		return helpers.ToApiError(err)
	}

	return helpers.SuccessResponse(c)
}

// UpdateReadme godoc
// @Summary      Update Readme
// @Description  Updating existing readme by id
// @Tags         Readmes
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        data formData dto.UpdateReadmeRequestDoc true "Readme update request"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/readmes/{readme} [patch]
func (rh *ReadmeHandl) UpdateReadme(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.FormValue("id")
	updates := make(map[string]any)
	title := c.FormValue("title")
	if title != "" {
		updates["title"] = title
	}
	form, err := c.MultipartForm()
	if err != nil {
		return helpers.InvalidRequest()
	}
	widgetsData := c.FormValue("widgets")
	if widgetsData != "" {
		widgets := make([]map[string]string, 0)
		if err := json.Unmarshal([]byte(widgetsData), &widgets); err != nil {
			return helpers.ToApiError(err)
		}
		updates["widgets"] = widgets
	}
	if links, ok := form.Value["links"]; ok {
		updates["links"] = links
	}
	if text, ok := form.Value["text"]; ok {
		updates["text"] = text
	}
	if render_order, ok := form.Value["render_order"]; ok {
		updates["render_order"] = render_order
	}

	if image, _ := c.FormFile("image"); image != nil {
		updates["image"] = image
	}
	req := dto.UpdateReadmeRequest{
		Updates: updates,
		Id:      id,
	}
	if errs := helpers.ValidateStruct(req, rh.Validator); len(errs) > 0 {
		return helpers.ValidationError(errs)
	}
	if err := rh.ReadmeServ.Update(ctx, req.Updates, req.Id); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// GetReadmeById godoc
// @Summary      Get Readme by ID
// @Description  Returns single readme by id
// @Tags         Readmes
// @Produce      json
// @Security     ApiKeyAuth
// @Param        readme path string true "Readme ID"
// @Success      200 {object} models.Readme "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/readmes/{readme} [get]
func (rh *ReadmeHandl) GetReadmeById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("readme")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	readme, err := rh.ReadmeServ.Get(ctx, id)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(readme)
}

// FetchReadmesByUser godoc
// @Summary      Fetch User Readmes
// @Description  Returns list of readmes for the authorized user
// @Tags         Readmes
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body query dto.PaginationRequest true "Pagination request"
// @Success      200 {array} dto.ReadmeResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/readmes [get]
func (rh *ReadmeHandl) FetchReadmesByUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	uid, err := utils.GetIdFromLocals(c.Locals("user"))
	if err != nil {
		return helpers.ToApiError(err)
	}
	req := dto.PaginationRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Query{}, rh.Validator); err != nil {
		return err
	}
	readmes, err := rh.ReadmeServ.FetchByUser(ctx, req.Amount, req.Page, uid)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(readmes)
}
