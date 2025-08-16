package server

import (
	"context"
	"errors"

	//"errors"
	"readmeow/internal/config"
	"readmeow/internal/delivery/handlers"
	//"readmeow/internal/delivery/handlers"
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
		ErrorHandler: errorHandler,
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
		validPaths := map[string]bool{
			"/api/auth/login":    true,
			"/api/auth/register": true,
			"/api/auth/verify":   true,
		}
		if validPaths[c.Path()] {
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

func errorHandler(c *fiber.Ctx, err error) error {
	var apiErr handlers.ApiErr
	if errors.As(err, &apiErr) {
		return c.Status(apiErr.Code).JSON(apiErr)
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "internal server error",
	})
}
