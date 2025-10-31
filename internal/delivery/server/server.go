package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"readmeow/internal/config"
	"readmeow/internal/delivery/middlewares"
	"readmeow/pkg/monitoring"
	"time"

	_ "readmeow/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	App    *fiber.App
	Metric *http.Server
	Config config.ServerConfig
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
	forgotPassword     = "/api/auth/forgot-password"
	resetPassword      = "/api/auth/reset-password"
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
	metricsMux.Handle("/metrics", promhttp.Handler())
	metrics := &http.Server{
		Addr:    ":" + scfg.MetricPort,
		Handler: metricsMux,
	}

	corsMiddleware := cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,DELETE,PATCH,OPTIONS",
		ExposeHeaders:    "Content-Length",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	})

	swaggerGroup := app.Group("/api/swagger")

	// validIps := map[string]struct{}{
	// 	"46.53.246.96": {},
	// 	"46.53.249.64": {},
	// }

	validAuthPaths := map[string]struct{}{
		login:              {},
		register:           {},
		verify:             {},
		newcode:            {},
		googleAuth:         {},
		googleAuthCallback: {},
		githubAuth:         {},
		githubAuthCallback: {},
		fetchTemplates:     {},
		fetchWidgets:       {},
		resetPassword:      {},
		forgotPassword:     {},
	}

	validAlreadyLoginPaths := map[string]struct{}{
		login:              {},
		register:           {},
		verify:             {},
		newcode:            {},
		googleAuth:         {},
		googleAuthCallback: {},
		githubAuth:         {},
		githubAuthCallback: {},
		resetPassword:      {},
		forgotPassword:     {},
	}

	swaggerGroup.Use(
		corsMiddleware,
		// middlewares.ValidIpsMiddleware(validIps),
	)

	swaggerGroup.Get("/*", swagger.HandlerDefault)

	writer := io.Writer(os.Stdout)
	if len(apcfg.LogPath) > 0 {
		logFile, err := os.OpenFile(apcfg.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(fmt.Errorf("failed to open log file: %w", err))
		}
		writer = io.MultiWriter(logFile, os.Stdout)
	}
	log := logger.New(logger.Config{
		Format:   "{\"time\":\"${time}\",\"method\":\"${method}\",\"path\":\"${path}\",\"status\":${status},\"latency\":\"${latency}\",\"type\":\"http\"}\n",
		TimeZone: "Europe/Minsk",
		Output:   writer,
	})

	app.Use(
		log,
		corsMiddleware,
		middlewares.AlreadyLoginCheck(validAlreadyLoginPaths),
		middlewares.AuthMiddleware(acfg, validAuthPaths),
		middlewares.RateLimiterMiddleware(scfg),
		middlewares.RequestTimeoutMiddleware(acfg.TokenTTL),
		middlewares.MetricsMiddleware(ps),
	)

	return &Server{App: app, Metric: metrics, Config: scfg}
}

func (s *Server) MustStart() {
	// if err := s.App.ListenTLS(s.Config.Host+":"+s.Config.Port, s.Config.CertFile, s.Config.KeyFile); err != nil {
	// 	panic("failed to start server: " + err.Error())
	// }
	if err := s.App.Listen(s.Config.Host + ":" + s.Config.Port); err != nil {
		panic("failed to start server: " + err.Error())
	}
}

func (s *Server) MustClose(ctx context.Context) {
	if err := s.Metric.Shutdown(ctx); err != nil {
		panic("failed to close metrics server" + err.Error())
	}
	if err := s.App.ShutdownWithContext(ctx); err != nil {
		panic("failed to close server" + err.Error())
	}
}
