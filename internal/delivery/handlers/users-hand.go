package handlers

import (
	"fmt"
	"readmeow/internal/delivery/handlers/helpers"
	"readmeow/internal/domain/services"
	"readmeow/internal/domain/utils"
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

// GetUser godoc
// @Summary      Get User
// @Description  Get user by ID
// @Tags         Users
// @Produce      json
// @Security     ApiKeyAuth
// @Param        user path string true "User ID"
// @Success      200 {object} dto.UserResponse "User data"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/users/{user} [get]
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

// UpdateUser godoc
// @Summary      Update User
// @Description  Update user's profile (nickname, avatar, etc.)
// @Tags         Users
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        data formData dto.UpdateUserRequestDoc true "Update user request"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/users [patch]
func (uh *UserHandl) Update(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := utils.GetIdFromLocals(c.Locals("user"))
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
		return helpers.ValidationError(errs)
	}
	if err := uh.UserServ.Update(ctx, req.Updates, id); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// DeleteUser godoc
// @Summary      Delete User
// @Description  Delete current user account
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body body dto.DeleteUserRequest true "Delete user request"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/users [delete]
func (uh *UserHandl) Delete(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.DeleteUserRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, uh.Validator); err != nil {
		return err
	}
	id, err := utils.GetIdFromLocals(c.Locals("user"))
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

// ChangeUserPassword godoc
// @Summary      Change User Password
// @Description  Change password for current user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body body dto.ChangePasswordRequest true "Change password request"
// @Success      200 {object} dto.SuccessResponse "Success response"
// @Failure      400 {object} helpers.ApiErr "Bad request"
// @Failure      401 {object} helpers.ApiErr "Unauthorized"
// @Failure      404 {object} helpers.ApiErr "Not found"
// @Failure      422 {object} helpers.ApiErr "Invalid JSON"
// @Failure      500 {object} helpers.ApiErr "Internal server error"
// @Router       /api/users/password [patch]
func (uh *UserHandl) ChangeUserPassword(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.ChangePasswordRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, uh.Validator); err != nil {
		return err
	}
	id, err := utils.GetIdFromLocals(c.Locals("user"))
	if err != nil {
		return helpers.ToApiError(err)
	}
	if err := uh.UserServ.ChangePassword(ctx, id, req.OldPasswrod, req.NewPassword); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}
