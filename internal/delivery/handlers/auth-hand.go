package handlers

import (
	"readmeow/internal/delivery/dto"
	"readmeow/internal/domain/services"
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
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := ah.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	if err := ah.AuthServ.Register(ctx, req.Email, req.Code); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to register user: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (ah *AuthHandl) VerifyEmail(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.VerifyRequest{}
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := ah.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	if err := ah.AuthServ.SendVerifyCode(ctx, req.Email, req.Login, req.Nickname, req.Password); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to send verification code: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func (ah *AuthHandl) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.LoginRequest{}
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := ah.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	loginResponce, err := ah.AuthServ.Login(ctx, req.Login, req.Password)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to login user: " + err.Error(),
		})
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
		Id:     loginResponce.Id.String(),
		Login:  loginResponce.Login,
		Avatar: loginResponce.Avatar,
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
		Path:     "/",
		SameSite: "Lax",
	}
	c.Cookie(cookie)
	return c.JSON(fiber.Map{
		"message": "succes",
	})
}

func (ah *AuthHandl) Profile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	cookie := c.Cookies("jwt")
	id, err := ah.AuthServ.GetId(ctx, cookie)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get user id: " + err.Error(),
		})
	}
	user, err := ah.UserServ.Get(ctx, id)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to get user: " + err.Error(),
		})
	}
	return c.JSON(user)
}

func (ah *AuthHandl) SendNewCode(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SendNewCodeRequest{}
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusUnprocessableEntity)
		return c.JSON(fiber.Map{
			"error": "failed to parse reqeust: " + err.Error(),
		})
	}
	if err := ah.Validator.Validate.Struct(req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "validation failed: " + err.Error(),
		})
	}
	if err := ah.AuthServ.SendNewCode(ctx, req.Email); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "failed to send new code: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
