package handlers

import (
	"readmeow/internal/delivery/dto"
	"readmeow/internal/domain/services"
	"readmeow/pkg/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ReadmeHandl struct {
	ReadmeServ services.ReadmeServ
	AuthServ   services.AuthServ
	Validator  *validator.Validator
}

func NewReadmeHandl(rs services.ReadmeServ, as services.AuthServ, v *validator.Validator) *ReadmeHandl {
	return &ReadmeHandl{
		ReadmeServ: rs,
		Validator:  v,
	}
}

func (rh *ReadmeHandl) CreateReadme(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.CreateReadmeRequest{}
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := rh.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	if err := rh.ReadmeServ.Create(ctx, req.TemplateId, req.OwnerId, req.Title, req.Order, req.Image, req.Text, req.Links, req.Widgets); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to create readme: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (rh *ReadmeHandl) DeleteReadme(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("readme")
	if err := uuid.Validate(id); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "invalid id: " + err.Error(),
		})
	}
	cookie := c.Cookies("jwt")
	uid, err := rh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get user id: " + err.Error(),
		})
	}
	if err := rh.ReadmeServ.Delete(ctx, id, uid); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to delete readme: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (rh *ReadmeHandl) UpdateReadme(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.UpdateReadmeRequest{}
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := rh.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	if err := rh.ReadmeServ.Update(ctx, req.Updates, req.Id); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to update readme: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (rh *ReadmeHandl) GetReadmeById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("readme")
	if err := uuid.Validate(id); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "invalid id: " + err.Error(),
		})
	}
	readme, err := rh.ReadmeServ.Get(ctx, id)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get readme: " + err.Error(),
		})
	}
	return c.JSON(readme)
}

func (rh *ReadmeHandl) FetchReadmesByUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.PaginationRequest{}
	if err := c.QueryParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := rh.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	cookie := c.Cookies("jwt")
	uid, err := rh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get user id: " + err.Error(),
		})
	}
	readmes, err := rh.ReadmeServ.FetchByUser(ctx, req.Amount, req.Page, uid)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to fetch readmes: " + err.Error(),
		})
	}
	return c.JSON(readmes)
}
