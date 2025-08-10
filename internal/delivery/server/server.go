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
	App *fiber.App
}

func NewServer(scfg config.ServerConfig, acfg config.AuthConfig) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(int(time.Second) * scfg.ReadTimeout),
		WriteTimeout: time.Duration(int(time.Second) * scfg.WriteTimeout),
		IdleTimeout:  time.Duration(int(time.Second) * scfg.IdleTimeout),
	})
	app.Use(
		cors.New(cors.Config{}),
		authMiddleware(acfg),
		requestTimeoutMiddleware(time.Duration(scfg.RequestTimeout*int(time.Second))),
	)

	return &Server{App: app}
}

func (s *Server) MustClose(ctx context.Context) {
	if err := s.App.ShutdownWithContext(ctx); err != nil {
		panic(err)
	}
}

func authMiddleware(acfg config.AuthConfig) fiber.Handler {
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
