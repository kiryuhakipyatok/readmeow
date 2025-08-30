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

// Register godoc
// @Summary Register
// @Description Register a verified user with email and verification code
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "Register Request"
// @Success 200 {object} dto.IdResponse "Success response"
// @Failure 400 {object} helpers.ApiErr "Bad request"
// @Failure 404 {object} helpers.ApiErr "Not found"
// @Failure 409 {object} helpers.ApiErr "Already exists"
// @Failure 422 {object} helpers.ApiErr "Invalid JSON"
// @Failure 500 {object} helpers.ApiErr "Internal server error"
// @Router /api/auth/register [post]
func (ah *AuthHandl) Register(c *fiber.Ctx) error {
	ctx := c.UserContext()
	if c.Cookies("jwt") != "" {
		return helpers.AlreadyLoggined(c)
	}
	req := dto.RegisterRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	id, err := ah.AuthServ.Register(ctx, req.Email, req.Code)
	if err != nil {
		return helpers.ToApiError(err)
	}
	idResp := dto.IdResponse{
		Id:      id,
		Message: "user registered successfully",
	}
	return c.JSON(idResp)
}

// VerifyEmail godoc
// @Summary Verify Email
// @Description Verifying a user by sending a verification code
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.VerifyRequest true "Verification Request"
// @Success 200 {object} dto.SuccessResponse "Success response"
// @Failure 400 {object} helpers.ApiErr "Bad request"
// @Failure 404 {object} helpers.ApiErr "Not found"
// @Failure 409 {object} helpers.ApiErr "Already exists"
// @Failure 422 {object} helpers.ApiErr "Invalid JSON"
// @Failure 500 {object} helpers.ApiErr "Internal server error"
// @Router /api/auth/verify [post]
func (ah *AuthHandl) VerifyEmail(c *fiber.Ctx) error {
	ctx := c.UserContext()
	if c.Cookies("jwt") != "" {
		return helpers.AlreadyLoggined(c)
	}
	req := dto.VerifyRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	if err := ah.AuthServ.SendVerifyCode(ctx, req.Email, req.Login, req.Nickname, req.Password); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// Login godoc
// @Summary Login
// @Description Log in a user with login and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "Login Request"
// @Success 200 {object} dto.LoginResponse "Login response"
// @Failure 400 {object} helpers.ApiErr "Bad request"
// @Failure 404 {object} helpers.ApiErr "Not found"
// @Failure 409 {object} helpers.ApiErr "Already exists"
// @Failure 422 {object} helpers.ApiErr "Invalid JSON"
// @Failure 500 {object} helpers.ApiErr "Internal server error"
// @Router /api/auth/login [post]
func (ah *AuthHandl) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()
	if c.Cookies("jwt") != "" {
		return helpers.AlreadyLoggined(c)
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

// Logout godoc
// @Summary Logout
// @Description Logout a user
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.SuccessResponse "Success response"
// @Security ApiKeyAuth
// @Router /api/auth/logout [get]
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

// Profile godoc
// @Summary Profile
// @Description Fetching the logged-in user's info
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.UserResponse "User info response"
// @Failure 404 {object} helpers.ApiErr "Not found"
// @Failure 500 {object} helpers.ApiErr "Internal server error"
// @Security ApiKeyAuth
// @Router /api/auth/profile [get]
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

// SendNewCode godoc
// @Summary Send New Code
// @Description Sending a new code for user verification
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.SendNewCodeRequest true "SendNewCodeRequest Request"
// @Success 200 {object} dto.SuccessResponse "Success response"
// @Failure 400 {object} helpers.ApiErr "Bad request"
// @Failure 404 {object} helpers.ApiErr "Not found"
// @Failure 409 {object} helpers.ApiErr "Already exists"
// @Failure 422 {object} helpers.ApiErr "Invalid JSON"
// @Failure 500 {object} helpers.ApiErr "Internal server error"
// @Router /api/auth/newcode [post]
func (ah *AuthHandl) SendNewCode(c *fiber.Ctx) error {
	ctx := c.UserContext()
	if c.Cookies("jwt") != "" {
		return helpers.AlreadyLoggined(c)
	}
	req := dto.SendNewCodeRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	if err := ah.AuthServ.SendNewCode(ctx, req.Email); err != nil {
		return helpers.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}
