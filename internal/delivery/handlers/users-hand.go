package handlers

import (
	"readmeow/internal/delivery/dto"
	"readmeow/internal/domain/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type UserHandl struct {
	UserServ  services.UserServ
	Validator *validator.Validate
}

func NewUserHandl(us services.UserServ, v *validator.Validate) *UserHandl {
	return &UserHandl{
		UserServ:  us,
		Validator: v,
	}
}

func (uh *UserHandl) Update(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.UpdateUserRequest{}
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := uh.Validator.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	if err := uh.UserServ.Update(ctx, req.Updates, req.Id); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to update user: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (uh *UserHandl) Delete(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.DeleteUserRequest{}
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := uh.Validator.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	if err := uh.UserServ.Delete(ctx, req.Id, req.Password); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to delete user: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (uh *UserHandl) ChangeUserPassword(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.ChangePasswordRequest{}
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := uh.Validator.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	if err := uh.UserServ.ChangePassword(ctx, req.Id, req.OldPasswrod, req.NewPassword); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to change user password: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
