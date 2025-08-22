package server

import (
	"context"
	"errors"
	"readmeow/internal/config"
	"readmeow/internal/delivery/handlers/helpers"
	"readmeow/internal/delivery/ratelimiter"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
)

type Server struct {
	App *fiber.App
}

func NewServer(scfg config.ServerConfig, acfg config.AuthConfig, apcfg config.AppConfig) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(scfg.ReadTimeout),
		WriteTimeout: time.Duration(scfg.WriteTimeout),
		IdleTimeout:  time.Duration(scfg.IdleTimeout),
		AppName:      apcfg.Name,
		//ErrorHandler: errorHandler,
	})
	app.Use(
		cors.New(cors.Config{}),
		authMiddleware(acfg),
		rateLimiterMiddleware(scfg),
		requestTimeoutMiddleware(acfg.TokenTTL),
	)

	return &Server{App: app}
}

func (s *Server) MustClose(ctx context.Context) {
	if err := s.App.ShutdownWithContext(ctx); err != nil {
		panic("failed to close server" + err.Error())
	}
}

const (
	login    = "/api/auth/login"
	register = "/api/auth/register"
	verify   = "/api/auth/verify"
)

func authMiddleware(acfg config.AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		validPaths := map[string]bool{
			login:    true,
			register: true,
			verify:   true,
		}
		if validPaths[c.Path()] {
			return c.Next()
		}
		return jwtware.New(jwtware.Config{
			SigningKey:  []byte(acfg.Secret),
			TokenLookup: "cookie:jwt",
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "unauthorized",
				})
			},
		})(c)
	}
}

func requestTimeoutMiddleware(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), timeout)
		defer cancel()
		c.SetUserContext(ctx)
		err := c.Next()
		if ctx.Err() == context.DeadlineExceeded {
			return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{
				"error": "request timeout",
			})
		}
		return err
	}
}

func errorHandler(c *fiber.Ctx, err error) error {
	var apiErr helpers.ApiErr
	if errors.As(err, &apiErr) {
		return c.Status(apiErr.Code).JSON(apiErr)
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "internal server error",
	})
}

func rateLimiterMiddleware(scfg config.ServerConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		limiter := ratelimiter.RateLimit(ip, scfg.RateLimit, scfg.Burst)
		if !limiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "too many requests",
			})
		}
		return c.Next()
	}
}
