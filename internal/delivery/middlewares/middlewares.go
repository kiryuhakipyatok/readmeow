package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"readmeow/internal/config"
	"readmeow/internal/delivery/apierr"
	"readmeow/internal/delivery/ratelimiter"
	"readmeow/pkg/monitoring"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
)

func AuthMiddleware(acfg config.AuthConfig, valid map[string]bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if valid[c.Path()] {
			return c.Next()
		}
		return jwtware.New(jwtware.Config{
			SigningKey:  []byte(acfg.Secret),
			TokenLookup: "cookie:jwt",
			SuccessHandler: func(c *fiber.Ctx) error {
				token, ok := c.Locals("user").(*jwt.Token)
				if !ok {
					return apierr.InvalidRequest()
				}
				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					return apierr.InvalidRequest()
				}
				userId, ok := claims["sub"].(string)
				if !ok {
					return apierr.InvalidRequest()
				}
				c.Locals("userId", userId)
				return c.Next()
			},
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				return apierr.Unauthorized()
			},
		})(c)
	}
}

func MetricsMiddleware(ps *monitoring.PrometheusSetup) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()
		statusCode := c.Response().StatusCode()
		if err != nil {
			var (
				statusCode int
				msg        string
			)
			switch err := err.(type) {
			case *fiber.Error:
				statusCode = err.Code
				msg = err.Message
			case apierr.ApiErr:
				statusCode = err.Code
				byteMsg, jerr := json.Marshal(err.Message)
				if jerr != nil {
					return apierr.InternalServerError()
				}
				msg = string(byteMsg)
			default:
				return apierr.InternalServerError()
			}
			ps.HTTPErrorTotal.WithLabelValues(c.Route().Path, c.Method(), strconv.Itoa(statusCode), msg).Inc()
		}
		ps.HTTPRequestsTotal.WithLabelValues(c.Route().Path, c.Method()).Inc()
		ps.HTTPRequestDuration.WithLabelValues(c.Route().Path, c.Method(), strconv.Itoa(statusCode)).Observe(duration)
		return err
	}
}

func RequestTimeoutMiddleware(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), timeout)
		defer cancel()
		c.SetUserContext(ctx)
		err := c.Next()
		if ctx.Err() == context.DeadlineExceeded {
			return apierr.RequestTimeout()
		}
		return err
	}
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(fe)
	}
	var apiErr apierr.ApiErr
	if errors.As(err, &apiErr) {
		return c.Status(apiErr.Code).JSON(apiErr)
	}
	return apierr.InternalServerError()
}

func RateLimiterMiddleware(scfg config.ServerConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		limiter := ratelimiter.RateLimit(ip, scfg.RateLimit, scfg.Burst)
		if !limiter.Allow() {
			return apierr.TooManyRequests()
		}
		return c.Next()
	}
}

func ValidIpsMiddleware(validIps map[string]bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if _, ok := validIps[c.IP()]; !ok {
			return apierr.Forbidden()
		}
		return c.Next()
	}
}

func AlreadyLoginCheck(valid map[string]bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !valid[c.Path()] {
			return c.Next()
		}
		if c.Cookies("jwt") != "" {
			return apierr.AlreadyLoggined()
		}
		return c.Next()
	}
}
