package server

import (
	"context"
	"readmeow/internal/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
)

type Server struct {
	*fiber.App
}

func NewServer(scfg *config.ServerConfig, acfg *config.AuthConfig) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(scfg.ReadTimeout),
		WriteTimeout: time.Duration(scfg.WriteTimeout),
		IdleTimeout:  time.Duration(scfg.IdleTimeout),
	})
	app.Use(
		cors.New(cors.Config{}),
		AuthMiddleware(acfg),
		RequestTimeoutMiddleware(time.Duration(scfg.RequestTimeout)),
	)

	return &Server{app}
}

func (s *Server) Close(ctx context.Context) {
	if err := s.App.ShutdownWithContext(ctx); err != nil {
		panic(err)
	}
}

func AuthMiddleware(acfg *config.AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/api/auth/login" || c.Path() == "/api/auth/register" {
			return c.Next()
		}
		return jwtware.New(jwtware.Config{
			SigningKey:  []byte(acfg.Secret),
			TokenLookup: "cookie:jwt",
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				c.Status(fiber.StatusUnauthorized)
				return c.JSON(fiber.Map{
					"message": "unauthorized",
				})
			},
		})(c)
	}
}

func RequestTimeoutMiddleware(timeout time.Duration) fiber.Handler {
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
