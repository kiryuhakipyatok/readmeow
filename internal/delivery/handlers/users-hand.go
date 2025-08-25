package handlers

import (
	"fmt"
	"readmeow/internal/delivery/handlers/helpers"
	"readmeow/internal/domain/services"
	"readmeow/internal/dto"
	"readmeow/pkg/validator"
	"time"

	"github.com/gofiber/fiber/v2"
)

type UserHandl struct {
	UserServ  services.UserServ
	AuthServ  services.AuthServ
	Validator *validator.Validator
}

func NewUserHandl(us services.UserServ, as services.AuthServ, v *validator.Validator) *UserHandl {
	return &UserHandl{
		UserServ:  us,
		AuthServ:  as,
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

	cookie := c.Cookies("jwt")
	id, err := uh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	updates := make(map[string]any)
	nickname := c.FormValue("nickname")
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if image, _ := c.FormFile("avatar"); image != nil {
		updates["avatar"] = image
	}
	req := dto.UpdateUserRequest{
		Updates: updates,
		Id:      id,
	}
	fmt.Println(req)
	if errs := helpers.ValidateStruct(req, uh.Validator); len(errs) > 0 {
		return err
	}
	if err := uh.UserServ.Update(ctx, req.Updates, id); err != nil {
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
	cookie := c.Cookies("jwt")
	id, err := uh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := uh.UserServ.Delete(ctx, id, req.Password); err != nil {
		return helpers.ToApiError(err)
	}
	newCookie := &fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		HTTPOnly: true,
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
		Path:     "/",
		SameSite: "Lax",
	}
	c.Cookie(newCookie)
	return helpers.SuccessResponse(c)
}

func (uh *UserHandl) ChangeUserPassword(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.ChangePasswordRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, uh.Validator); err != nil {
		return err
	}
	cookie := c.Cookies("jwt")
	id, err := uh.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := uh.UserServ.ChangePassword(ctx, id, req.OldPasswrod, req.NewPassword); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}
