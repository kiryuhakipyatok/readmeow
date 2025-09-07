package handlers

import (
	"encoding/json"
	"readmeow/internal/delivery/apierr"
	"readmeow/internal/delivery/handlers/helpers"
	"readmeow/internal/delivery/oauth"
	"readmeow/internal/domain/services"
	"readmeow/internal/dto"
	"readmeow/pkg/validator"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AuthHandl struct {
	AuthServ    services.AuthServ
	UserServ    services.UserServ
	OAuthConfig oauth.OAuthConfig
	Validator   *validator.Validator
}

func NewAuthHandle(as services.AuthServ, us services.UserServ, oc oauth.OAuthConfig, v *validator.Validator) *AuthHandl {
	return &AuthHandl{
		AuthServ:    as,
		UserServ:    us,
		OAuthConfig: oc,
		Validator:   v,
	}
}

// Register godoc
// @Summary Register
// @Description Register a verified user with email and verification code
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "Register Request"
// @Success 200 {object} dto.SuccessResponse "Success response"
// @Failure 400 {object} apierr.ApiErr "Bad request"
// @Failure 404 {object} apierr.ApiErr "Not found"
// @Failure 409 {object} apierr.ApiErr "Already exists"
// @Failure 422 {object} apierr.ApiErr "Invalid JSON"
// @Failure 500 {object} apierr.ApiErr "Internal server error"
// @Router /api/auth/register [post]
func (ah *AuthHandl) Register(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.RegisterRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	id, err := ah.AuthServ.Register(ctx, req.Email, req.Code)
	if err != nil {
		return apierr.ToApiError(err)
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
// @Failure 400 {object} apierr.ApiErr "Bad request"
// @Failure 404 {object} apierr.ApiErr "Not found"
// @Failure 409 {object} apierr.ApiErr "Already exists"
// @Failure 422 {object} apierr.ApiErr "Invalid JSON"
// @Failure 500 {object} apierr.ApiErr "Internal server error"
// @Router /api/auth/verify [post]
func (ah *AuthHandl) VerifyEmail(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.VerifyRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	if err := ah.AuthServ.SendVerifyCode(ctx, req.Email, req.Login, req.Nickname, req.Password); err != nil {
		return apierr.ToApiError(err)
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
// @Failure 400 {object} apierr.ApiErr "Bad request"
// @Failure 404 {object} apierr.ApiErr "Not found"
// @Failure 409 {object} apierr.ApiErr "Already exists"
// @Failure 422 {object} apierr.ApiErr "Invalid JSON"
// @Failure 500 {object} apierr.ApiErr "Internal server error"
// @Router /api/auth/login [post]
func (ah *AuthHandl) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.LoginRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	loginResponce, err := ah.AuthServ.Login(ctx, req.Login, req.Password)
	if err != nil {
		return apierr.ToApiError(err)
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
		Id:       loginResponce.Id,
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
// @Failure 404 {object} apierr.ApiErr "Not found"
// @Failure 500 {object} apierr.ApiErr "Internal server error"
// @Security ApiKeyAuth
// @Router /api/auth/profile [get]
func (ah *AuthHandl) Profile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := c.Locals("userId").(string)
	user, err := ah.UserServ.Get(ctx, id, true)
	if err != nil {
		return apierr.ToApiError(err)
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
// @Failure 400 {object} apierr.ApiErr "Bad request"
// @Failure 404 {object} apierr.ApiErr "Not found"
// @Failure 409 {object} apierr.ApiErr "Already exists"
// @Failure 422 {object} apierr.ApiErr "Invalid JSON"
// @Failure 500 {object} apierr.ApiErr "Internal server error"
// @Router /api/auth/newcode [post]
func (ah *AuthHandl) SendNewCode(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := dto.SendNewCodeRequest{}
	if err := helpers.ParseAndValidateRequest(c, &req, helpers.Body{}, ah.Validator); err != nil {
		return err
	}
	if err := ah.AuthServ.SendNewCode(ctx, req.Email); err != nil {
		return apierr.ToApiError(err)
	}
	return helpers.SuccessResponse(c)
}

// GoogleOAuth godoc
// @Summary Login via Google
// @Description Start Google OAuth login flow. User will be redirected to Google for authentication, then back to your app. After successful login, the client will receive dto.LoginResponse.
// @Tags Auth
// @Produce json
// @Success 307 "Redirect to Google OAuth"
// @Failure 400 {object} apierr.ApiErr "Bad request"
// @Failure 500 {object} apierr.ApiErr "Internal server error"
// @Router /api/auth/google [get]
func (ah *AuthHandl) GoogleOAuth(c *fiber.Ctx) error {
	state := uuid.New().String()
	exp := time.Now().Add(ah.OAuthConfig.StateExpire)
	stateCookie := &fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Expires:  exp,
		MaxAge:   int(time.Until(exp).Seconds()),
		SameSite: "Lax",
	}
	c.Cookie(stateCookie)
	url := ah.OAuthConfig.GoogleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

// GitHubOAuth godoc
// @Summary Login via GitHub
// @Description Start GitHub OAuth login flow. User will be redirected to Google for authentication, then back to your app. After successful login, the client will receive dto.LoginResponse.
// @Tags Auth
// @Produce json
// @Success 307 "Redirect to GitHub OAuth"
// @Failure 400 {object} apierr.ApiErr "Bad request"
// @Failure 500 {object} apierr.ApiErr "Internal server error"
// @Router /api/auth/github [get]
func (ah *AuthHandl) GitHubOAuth(c *fiber.Ctx) error {
	state := uuid.New().String()
	exp := time.Now().Add(ah.OAuthConfig.StateExpire)
	stateCookie := &fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Expires:  exp,
		MaxAge:   int(time.Until(exp).Seconds()),
		SameSite: "Lax",
	}
	c.Cookie(stateCookie)
	url := ah.OAuthConfig.GithubOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

const (
	github = "github"
	google = "google"
)

func (ah *AuthHandl) GoogleOAthCallback(c *fiber.Ctx) error {
	ctx := c.UserContext()
	state := c.Query("state")
	defer func() {
		s := &fiber.Cookie{
			Name:     "oauth_state",
			Value:    "",
			HTTPOnly: true,
			Expires:  time.Now().Add(-time.Hour),
			MaxAge:   -1,
			SameSite: "Lax",
		}
		c.Cookie(s)
	}()
	if c.Cookies("oauth_state") != state {
		return helpers.InvalidState(c)
	}
	code := c.Query("code")
	token, err := ah.OAuthConfig.GoogleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		return apierr.ToApiError(err)
	}
	client := ah.OAuthConfig.GoogleOAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return apierr.ToApiError(err)
	}
	defer resp.Body.Close()
	oauthReq := dto.GoogleOAuthRequest{}
	if err := json.NewDecoder(resp.Body).Decode(&oauthReq); err != nil {
		return apierr.ToApiError(err)
	}

	loginResponce, err := ah.AuthServ.OAuthLogin(ctx, oauthReq.Name, oauthReq.Picture, oauthReq.Email, oauthReq.Id, google)
	if err != nil {
		return apierr.ToApiError(err)
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
		Id:       loginResponce.Id,
		Nickname: loginResponce.Nickname,
		Avatar:   loginResponce.Avatar,
	}
	return c.JSON(responce)
}

func (ah *AuthHandl) GitHubOAuthCallback(c *fiber.Ctx) error {
	ctx := c.UserContext()
	state := c.Query("state")
	defer func() {
		s := &fiber.Cookie{
			Name:     "oauth_state",
			Value:    "",
			HTTPOnly: true,
			Expires:  time.Now().Add(-time.Hour),
			MaxAge:   -1,
			SameSite: "Lax",
		}
		c.Cookie(s)
	}()
	if c.Cookies("oauth_state") != state {
		return helpers.InvalidState(c)
	}
	code := c.Query("code")
	token, err := ah.OAuthConfig.GithubOAuthConfig.Exchange(ctx, code)
	if err != nil {
		return apierr.ToApiError(err)
	}
	client := ah.OAuthConfig.GithubOAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return apierr.ToApiError(err)
	}
	defer resp.Body.Close()
	oauthReq := dto.GitHubOAuthRequest{}
	if err := json.NewDecoder(resp.Body).Decode(&oauthReq); err != nil {
		return apierr.ToApiError(err)
	}
	emailResp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return apierr.ToApiError(err)
	}
	defer emailResp.Body.Close()
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
		return apierr.ToApiError(err)
	}
	for _, e := range emails {
		if e.Verified && e.Primary {
			oauthReq.Email = e.Email
			break
		}
	}
	pid := strconv.FormatInt(oauthReq.Id, 10)
	loginResponce, err := ah.AuthServ.OAuthLogin(ctx, oauthReq.Login, oauthReq.Avatar, oauthReq.Email, pid, github)
	if err != nil {
		return apierr.ToApiError(err)
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
		Id:       loginResponce.Id,
		Nickname: loginResponce.Nickname,
		Avatar:   loginResponce.Avatar,
	}
	return c.JSON(responce)
}
