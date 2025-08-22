package handlers

import (
	"encoding/json"
	"readmeow/internal/delivery/handlers/helpers"
	"readmeow/internal/domain/services"
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

func (rh *ReadmeHandl) CreateReadme(c *fiber.Ctx) error {
	ctx := c.UserContext()
	cookie := c.Cookies("jwt")

	req := dto.CreateReadmeRequest{}
	uid, err := rh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return err
	}

	req.TemplateId = c.FormValue("template_id")

	req.Title = c.FormValue("title")
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	renderOrder := form.Value["render_order"]
	req.RenderOrder = renderOrder
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
			return err
		}
		req.Widgets = widgets
	}
	image, err := c.FormFile("image")
	if err != nil && err.Error() != "there is no uploaded file associated with the given key" {
		return err
	}
	if image != nil {
		req.Image = image
	}
	if err := rh.Validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := rh.ReadmeServ.Create(ctx, req.TemplateId, uid, req.Title, req.Image, req.Text, req.Links, req.RenderOrder, req.Widgets); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

func (rh *ReadmeHandl) DeleteReadme(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("readme")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := rh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := rh.ReadmeServ.Delete(ctx, id, uid); err != nil {
		return helpers.ToApiError(err)
	}

	return helpers.SuccessResponse(c)
}

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
		return helpers.ToApiError(err)
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

	image, err := c.FormFile("image")
	if err != nil && err.Error() != "there is no uploaded file associated with the given key" {
		return err
	}
	if image != nil {
		updates["image"] = image
	}
	req := dto.UpdateReadmeRequest{
		Updates: updates,
		Id:      id,
	}
	if err := rh.Validator.Validate.Struct(req); err != nil {
		return helpers.InvalidRequest()
	}
	if err := rh.ReadmeServ.Update(ctx, req.Updates, req.Id); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

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

func (rh *ReadmeHandl) FetchReadmesByUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.PaginationRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Query{}, rh.Validator); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := rh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	readmes, err := rh.ReadmeServ.FetchByUser(ctx, req.Amount, req.Page, uid)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(readmes)
}
