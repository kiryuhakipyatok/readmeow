package handlers

import (
	"readmeow/internal/delivery/dto"
	"readmeow/internal/domain/services"
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
	req := dto.CreateReadmeRequest{}
	if err := ParseAndValidateRequest(c, &req, Body{}, rh.Validator); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := rh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return ToApiError(err)
	}
	if err := rh.ReadmeServ.Create(ctx, req.TemplateId, uid, req.Title, req.Image, req.Text, req.Links, req.Order, req.Widgets); err != nil {
		return ToApiError(err)
	}
	return SuccessResponse(c)
}

func (rh *ReadmeHandl) DeleteReadme(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("readme")
	if err := ValidateId(c, id); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := rh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return ToApiError(err)
	}
	if err := rh.ReadmeServ.Delete(ctx, id, uid); err != nil {
		return ToApiError(err)
	}

	return SuccessResponse(c)
}

func (rh *ReadmeHandl) UpdateReadme(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.UpdateReadmeRequest{}
	if err := ParseAndValidateRequest(c, &req, Body{}, rh.Validator); err != nil {
		return err
	}
	if err := rh.ReadmeServ.Update(ctx, req.Updates, req.Id); err != nil {
		return ToApiError(err)
	}
	return SuccessResponse(c)
}

func (rh *ReadmeHandl) GetReadmeById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("readme")
	if err := ValidateId(c, id); err != nil {
		return err
	}
	readme, err := rh.ReadmeServ.Get(ctx, id)
	if err != nil {
		return ToApiError(err)
	}
	return c.JSON(readme)
}

func (rh *ReadmeHandl) FetchReadmesByUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.PaginationRequest{}
	if err := ParseAndValidateRequest(c, &req, Query{}, rh.Validator); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	uid, err := rh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return ToApiError(err)
	}
	readmes, err := rh.ReadmeServ.FetchByUser(ctx, req.Amount, req.Page, uid)
	if err != nil {
		return ToApiError(err)
	}
	return c.JSON(readmes)
}
