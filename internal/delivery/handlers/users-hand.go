package handlers

import (
	"readmeow/internal/delivery/dto"
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
	if err := ValidateId(c, id); err != nil {
		return err
	}
	user, err := uh.UserServ.Get(ctx, id)
	if err != nil {
		return ToApiError(err)
	}
	return c.JSON(user)
}

func (uh *UserHandl) Update(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.UpdateUserRequest{}
	if err := ParseAndValidateRequest(c, &req, Body{}, uh.Validator); err != nil {
		return err
	}
	if err := uh.UserServ.Update(ctx, req.Updates, req.Id); err != nil {
		return ToApiError(err)
	}
	return SuccessResponse(c)
}

func (uh *UserHandl) Delete(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.DeleteUserRequest{}
	if err := ParseAndValidateRequest(c, &req, Body{}, uh.Validator); err != nil {
		return err
	}
	if err := uh.UserServ.Delete(ctx, req.Id, req.Password); err != nil {
		return ToApiError(err)
	}
	return SuccessResponse(c)
}

func (uh *UserHandl) ChangeUserPassword(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.ChangePasswordRequest{}
	if err := ParseAndValidateRequest(c, &req, Body{}, uh.Validator); err != nil {
		return err
	}
	if err := uh.UserServ.ChangePassword(ctx, req.Id, req.OldPasswrod, req.NewPassword); err != nil {
		return ToApiError(err)
	}
	return SuccessResponse(c)
}
