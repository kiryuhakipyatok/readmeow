package server

import (
	"context"
	"net/http"
	"readmeow/internal/config"
	"readmeow/internal/delivery/middlewares"
	"readmeow/pkg/monitoring"
	"time"

	_ "readmeow/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	App    *fiber.App
	Metric *http.Server
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

func NewServer(scfg config.ServerConfig, acfg config.AuthConfig, apcfg config.AppConfig, ps *monitoring.PrometheusSetup) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(scfg.ReadTimeout),
		WriteTimeout: time.Duration(scfg.WriteTimeout),
		IdleTimeout:  time.Duration(scfg.IdleTimeout),
		AppName:      apcfg.Name,
		ErrorHandler: middlewares.ErrorHandler,
	})
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.HandlerFor(ps.Registry, promhttp.HandlerOpts{}))
	metrics := &http.Server{
		Addr:    ":" + scfg.MetricPort,
		Handler: metricsMux,
	}

	corsMiddleware := cors.New(cors.Config{
		AllowOrigins:     "localhost:3000",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,DELETE,PATCH",
		ExposeHeaders:    "Content-Length",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	})

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

	validAlreadyLoginPaths := map[string]bool{
		login:              true,
		register:           true,
		verify:             true,
		newcode:            true,
		googleAuth:         true,
		googleAuthCallback: true,
		githubAuth:         true,
		githubAuthCallback: true,
	}

	swaggerGroup.Use(
		corsMiddleware,
		// validIpsMiddleware(validIps),
	)

	swaggerGroup.Get("/*", swagger.HandlerDefault)

	app.Use(
		corsMiddleware,
		middlewares.AuthMiddleware(acfg, validAuthPaths),
		middlewares.AlreadyLoginCheck(validAlreadyLoginPaths),
		middlewares.RateLimiterMiddleware(scfg),
		middlewares.RequestTimeoutMiddleware(acfg.TokenTTL),
		middlewares.MetricsMiddleware(ps),
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
