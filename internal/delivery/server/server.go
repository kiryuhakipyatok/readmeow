package server

import (
	"context"
	"errors"
	"net/http"
	"readmeow/internal/config"
	"readmeow/internal/delivery/apierr"
	"readmeow/internal/delivery/ratelimiter"
	"readmeow/pkg/monitoring"
	"strconv"
	"time"

	_ "readmeow/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/gofiber/swagger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	App    *fiber.App
	Metric *http.Server
}

func NewServer(scfg config.ServerConfig, acfg config.AuthConfig, apcfg config.AppConfig, ps *monitoring.PrometheusSetup) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(scfg.ReadTimeout),
		WriteTimeout: time.Duration(scfg.WriteTimeout),
		IdleTimeout:  time.Duration(scfg.IdleTimeout),
		AppName:      apcfg.Name,
		ErrorHandler: errorHandler,
	})
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.HandlerFor(ps.Registry, promhttp.HandlerOpts{}))
	metrics := &http.Server{
		Addr:    ":" + scfg.MetricPort,
		Handler: metricsMux,
	}

	corsMiddleware := cors.New(cors.Config{})

	swaggerGroup := app.Group("/api/swagger")

	// validIps := map[string]bool{
	// 	"172.21.0.1": true,
	// 	//	"000.00.0.0": true,
	// }

	validAuthPaths := map[string]bool{
		login:              true,
		register:           true,
		verify:             true,
		newcode:            true,
		googleAuth:         true,
		googleAuthCallback: true,
		githubAuth:         true,
		githubAuthCallback: true,
		fetchTemplates:     true,
		fetchWidgets:       true,
	}

	swaggerGroup.Use(
		corsMiddleware,
		// validIpsMiddleware(validIps),
	)

	swaggerGroup.Get("/*", swagger.HandlerDefault)

	app.Use(
		corsMiddleware,
		authMiddleware(acfg, validAuthPaths),
		alreadyLoginCheck(validAuthPaths),
		rateLimiterMiddleware(scfg),
		requestTimeoutMiddleware(acfg.TokenTTL),
		metricsMiddleware(ps),
	)

	return &Server{App: app, Metric: metrics}
}

func (s *Server) MustClose(ctx context.Context) {
	if err := s.Metric.Shutdown(ctx); err != nil {
		panic("failed to close metrics server" + err.Error())
	}
	if err := s.App.ShutdownWithContext(ctx); err != nil {
		panic("failed to close server" + err.Error())
	}
}

const (
	login              = "/api/auth/login"
	register           = "/api/auth/register"
	verify             = "/api/auth/verify"
	newcode            = "/api/auth/newcode"
	googleAuth         = "/api/auth/google"
	googleAuthCallback = "/api/auth/google/callback"
	githubAuth         = "/api/auth/github"
	githubAuthCallback = "/api/auth/github/callback"
	fetchTemplates     = "/api/templates"
	fetchWidgets       = "/api/widgets"
)

func authMiddleware(acfg config.AuthConfig, valid map[string]bool) fiber.Handler {
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
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "invalid local data",
					})
				}
				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "invalid token data",
					})
				}
				userId, ok := claims["sub"].(string)
				if !ok {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "invalid claims data",
					})
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

func metricsMiddleware(ps *monitoring.PrometheusSetup) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()
		statusCode := c.Response().StatusCode()
		if err != nil {
			switch err := err.(type) {
			case *fiber.Error:
				statusCode = err.Code
				ps.HTTPErrorTotal.WithLabelValues(c.Route().Path, c.Method(), strconv.Itoa(statusCode), err.Message).Inc()
			case apierr.ApiErr:
				statusCode = err.Code
				ps.HTTPErrorTotal.WithLabelValues(c.Route().Path, c.Method(), strconv.Itoa(statusCode), err.Message.(string)).Inc()
			}

		}
		ps.HTTPRequestsTotal.WithLabelValues(c.Route().Path, c.Method()).Inc()
		ps.HTTPRequestDuration.WithLabelValues(c.Route().Path, c.Method(), strconv.Itoa(statusCode)).Observe(duration)
		return err
	}
}

func requestTimeoutMiddleware(timeout time.Duration) fiber.Handler {
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

func errorHandler(c *fiber.Ctx, err error) error {
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

func rateLimiterMiddleware(scfg config.ServerConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		limiter := ratelimiter.RateLimit(ip, scfg.RateLimit, scfg.Burst)
		if !limiter.Allow() {
			return apierr.TooManyRequests()
		}
		return c.Next()
	}
}

// func validIpsMiddleware(validIps map[string]bool) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		if _, ok := validIps[c.IP()]; !ok {
// 			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 				"message": "forbidden",
// 			})
// 		}
// 		return c.Next()
// 	}
// }

func alreadyLoginCheck(valid map[string]bool) fiber.Handler {
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
