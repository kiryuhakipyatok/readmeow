package handlers

import (
	"readmeow/internal/delivery/handlers/helpers"
	"readmeow/internal/domain/services"
	"readmeow/internal/dto"
	"readmeow/pkg/validator"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AuthHandl struct {
	AuthServ  services.AuthServ
	UserServ  services.UserServ
	Validator *validator.Validator
}

func NewAuthHandle(as services.AuthServ, us services.UserServ, v *validator.Validator) *AuthHandl {
	return &AuthHandl{
		AuthServ:  as,
		UserServ:  us,
		Validator: v,
	}
}

func (ah *AuthHandl) Register(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.RegisterRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	if err := ah.AuthServ.Register(ctx, req.Email, req.Code); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

func (ah *AuthHandl) VerifyEmail(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.VerifyRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	if err := ah.AuthServ.SendVerifyCode(ctx, req.Email, req.Login, req.Nickname, req.Password); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

func (ah *AuthHandl) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()
	if c.Cookies("jwt") != "" {
		return helpers.AlreadyExists()
	}
	req := dto.LoginRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	loginResponce, err := ah.AuthServ.Login(ctx, req.Login, req.Password)
	if err != nil {
		return helpers.ToApiError(err)
	}
	cookie := &fiber.Cookie{
		Name:     "jwt",
		Value:    loginResponce.JWT,
		HTTPOnly: true,
		Expires:  loginResponce.TTL,
		MaxAge:   int(time.Until(loginResponce.TTL).Seconds()),
		SameSite: "Lax",
	}
	c.Cookie(cookie)
	responce := dto.LoginResponse{
		Id:       loginResponce.Id.String(),
		Nickname: loginResponce.Nickname,
		Avatar:   loginResponce.Avatar,
	}
	return c.JSON(responce)
}

func (ah *AuthHandl) Logout(c *fiber.Ctx) error {
	cookie := &fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		HTTPOnly: true,
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
		SameSite: "Lax",
	}
	c.Cookie(cookie)
	return helpers.SuccessResponse(c)
}

func (ah *AuthHandl) Profile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	cookie := c.Cookies("jwt")
	id, err := ah.AuthServ.GetId(ctx, cookie)
	if err != nil {
		return helpers.ToApiError(err)
	}
	user, err := ah.UserServ.Get(ctx, id)
	if err != nil {
		return helpers.ToApiError(err)
	}
	return c.JSON(user)
}

func (ah *AuthHandl) SendNewCode(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SendNewCodeRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	if err := ah.AuthServ.SendNewCode(ctx, req.Email); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}
