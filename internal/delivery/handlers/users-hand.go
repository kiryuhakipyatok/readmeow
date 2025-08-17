package handlers

import (
	"readmeow/internal/delivery/dto"
	"readmeow/internal/delivery/handlers/helpers"
	"readmeow/internal/domain/services"
	"readmeow/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type UserHandl struct {
	UserServ  services.UserServ
	Validator *validator.Validator
}

func NewUserHandl(us services.UserServ, v *validator.Validator) *UserHandl {
	return &UserHandl{
		UserServ:  us,
		Validator: v,
	}
}

func (uh *UserHandl) GetUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Params("user")
	if err := helpers.ValidateId(c, id); err != nil {
		return err
	}
	user, err := uh.UserServ.Get(ctx, id)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(user)
}

func (uh *UserHandl) Update(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.UpdateUserRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, uh.Validator); err != nil {
		return err
	}
	if err := uh.UserServ.Update(ctx, req.Updates, req.Id); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

func (uh *UserHandl) Delete(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.DeleteUserRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, uh.Validator); err != nil {
		return err
	}
	if err := uh.UserServ.Delete(ctx, req.Id, req.Password); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

func (uh *UserHandl) ChangeUserPassword(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.ChangePasswordRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, uh.Validator); err != nil {
		return err
	}
	if err := uh.UserServ.ChangePassword(ctx, req.Id, req.OldPasswrod, req.NewPassword); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}
