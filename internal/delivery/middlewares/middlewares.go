package middlewares

import (
	"context"
	"errors"
	"fmt"
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

func AuthMiddleware(acfg config.AuthConfig, valid map[string]struct{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		jwtHand := jwtware.New(jwtware.Config{
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
		path := c.Path()
		// if (strings.HasPrefix(path, "/api/auth") && !strings.HasSuffix(path, "/logout") && !strings.HasSuffix(path, "/profile")) || strings.HasSuffix(path, "/widgets"){
		// 	return c.Next()
		// }
		if _, ok := valid[path]; ok {
			return c.Next()
		}

		// 	return c.Next()
		// }
		return jwtHand
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
			)
			switch err := err.(type) {
			case *fiber.Error:
				statusCode = err.Code
			case apierr.ApiErr:
				statusCode = err.Code
			default:
				return apierr.InternalServerError()
			}
			ps.HTTPErrorTotal.WithLabelValues(c.Route().Path, c.Method(), strconv.Itoa(statusCode)).Inc()
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
		var key string
		ip := c.IP()
		userAgent := c.Get("User-Agent")
		id, ok := c.Locals("userId").(string)
		if ok && len(id) > 0 {
			key = fmt.Sprintf("%s|%s|%s", ip, id, userAgent)
		} else {
			key = fmt.Sprintf("%s|%s", ip, userAgent)
		}
		// identifier := sha256.Sum224([]byte(key))
		limiter := ratelimiter.RateLimit(key, scfg.RateLimit, scfg.Burst)
		if !limiter.Allow() {
			return apierr.TooManyRequests()
		}
		return c.Next()
	}
}

func ValidIpsMiddleware(validIps map[string]struct{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Println(c.IP())
		if _, ok := validIps[c.IP()]; !ok {
			return apierr.Forbidden()
		}
		return c.Next()
	}
}

func AlreadyLoginCheck(valid map[string]struct{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		// if !strings.HasPrefix(path, "/api/auth") || strings.HasSuffix(path, "/logout") || strings.HasSuffix(path, "/profile") {
		// 	return c.Next()
		// }
		if _, ok := valid[path]; !ok {
			return c.Next()
		}
		if c.Cookies("jwt") != "" {
			return apierr.AlreadyLoggined()
		}
		return c.Next()
	}
}
