package app

import (
	"context"
	"net/smtp"
	"os"
	"os/signal"
	"readmeow/internal/config"
	"readmeow/internal/delivery/handlers"
	"readmeow/internal/delivery/oauth"
	"readmeow/internal/delivery/routes"
	"readmeow/internal/delivery/server"
	"readmeow/internal/domain/repositories"
	"readmeow/internal/domain/services"
	"readmeow/internal/email"
	"readmeow/internal/scheduler"
	"readmeow/pkg/cache"
	"readmeow/pkg/cloudstorage"
	"readmeow/pkg/logger"
	"readmeow/pkg/monitoring"
	"readmeow/pkg/search"
	stor "readmeow/pkg/storage"
	"readmeow/pkg/validator"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func Run() {
	if err := godotenv.Load(".env"); err != nil {
		panic("failed to load .env" + err.Error())
	}
	cfg := config.MustLoadConfig(os.Getenv("CONFIG_PATH"))

	log := logger.NewLogger(cfg.App)

	log.Log.Info("config loaded")

	validator := validator.NewValidator()

	storage := stor.MustConnect(cfg.Storage)
	defer func() {
		storage.Close()
		log.Log.Info("postgres closed")
	}()
	log.Log.Info("connected to postgres")

	cache := cache.MustConnect(cfg.Cache)
	defer func() {
		cache.MustClose()
		log.Log.Info("redis closed")
	}()
	log.Log.Info("connected to redis")

	search := search.MustConnect(cfg.Search)
	log.Log.Info("connected to elasticsearch")

	ps := monitoring.NewPrometheusSetup()
	log.Log.Info("prometheus setuped")

	server := server.NewServer(cfg.Server, cfg.Auth, cfg.App, ps)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.Server.CloseTimeout))
		defer cancel()
		server.MustClose(ctx)
		log.Log.Info("server closed")
	}()
	log.Log.Info("server created")

	smtpAuth := smtp.PlainAuth("", cfg.Email.Address, cfg.Email.Password, cfg.Email.SmtpAddress)

	cloudStorage := cloudstorage.MustConnect(cfg.CloudStorage)
	log.Log.Info("connected to cloudinary")

	userRepo := repositories.NewUserRepo(storage)
	widgetRepo := repositories.NewWidgetRepo(storage, cache, search)
	readmeRepo := repositories.NewReadmeStorage(storage)
	templateRepo := repositories.NewTemplateRepo(storage, cache, search)
	verificationRepo := repositories.NewVerificationRepo(storage)
	transactor := stor.NewTransactor(storage)
	emailSendler := email.NewEmailSender(smtpAuth, cfg.Email)
	oauthConf := oauth.NewOAuthConfig(cfg.OAuth)

	authServ := services.NewAuthServ(userRepo, verificationRepo, cloudStorage, transactor, emailSendler, log, cfg.Auth)
	readmeServ := services.NewReadmeServ(readmeRepo, userRepo, templateRepo, widgetRepo, transactor, cloudStorage, log)
	widgetServ := services.NewWidgetServ(widgetRepo, userRepo, cloudStorage, transactor, log)
	templateServ := services.NewTemplateServ(templateRepo, readmeRepo, userRepo, widgetRepo, transactor, cloudStorage, log)
	userServ := services.NewUserServ(userRepo, templateRepo, cloudStorage, transactor, log)

	authHandl := handlers.NewAuthHandle(authServ, userServ, oauthConf, validator)
	readmeHandl := handlers.NewReadmeHandl(readmeServ, authServ, validator)
	widgetHandl := handlers.NewWidgetHandl(widgetServ, authServ, validator)
	templateHandl := handlers.NewTemplateHandl(templateServ, authServ, validator)
	userHandl := handlers.NewUserHandl(userServ, authServ, validator)

	sheduler := scheduler.NewScheduler(widgetRepo, templateRepo, verificationRepo, cfg.Sheduler, cfg.Search, log)
	sheduler.Start()
	defer func() {
		sheduler.Stop()
		log.Log.Info("sheduler stopped")
	}()
	log.Log.Info("sheduler started")

	routConfig := routes.NewRoutConfig(server.App, userHandl, authHandl, templateHandl, readmeHandl, widgetHandl)
	routConfig.SetupRoutes()

	go func() {
		if err := server.App.Listen(cfg.Server.Host + ":" + cfg.Server.Port); err != nil {
			panic("failed to start server: " + err.Error())
		}
	}()
	go func() {
		if err := server.Metric.ListenAndServe(); err != nil {
			panic("failed to start server of metrics: " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Log.Info("app shutting down...")
}
