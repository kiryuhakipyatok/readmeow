package app

import (
	"context"
	"net/smtp"
	"os"
	"os/signal"
	"readmeow/internal/config"
	"readmeow/internal/delivery/handlers"
	"readmeow/internal/delivery/routs"
	"readmeow/internal/delivery/server"
	"readmeow/internal/domain/repositories"
	"readmeow/internal/domain/services"
	"readmeow/internal/email"
	"readmeow/internal/sheduler"
	"readmeow/pkg/cache"
	"readmeow/pkg/logger"
	"readmeow/pkg/search"
	stor "readmeow/pkg/storage"
	"readmeow/pkg/validator"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func Run() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	log := logger.NewLogger(os.Getenv("APP_ENV"))

	cfg := config.LoadConfig(os.Getenv("CONFIG_PATH"))
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

	app := server.NewServer(cfg.Server, cfg.Auth)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.Server.CloseTimeout))
		defer cancel()
		app.MustClose(ctx)
		log.Log.Info("server closed")
	}()
	log.Log.Info("server created")

	smtpAuth := smtp.PlainAuth("", cfg.Email.Address, cfg.Email.Password, cfg.Email.SmtpAddress)

	userRepo := repositories.NewUserRepo(storage)
	widgetRepo := repositories.NewWidgetRepo(storage, cache, search)
	readmeRepo := repositories.NewReadmeStorage(storage)
	templateRepo := repositories.NewTemplateRepo(storage, cache, search)
	verificationRepo := repositories.NewVerificationRepo(storage)
	transactor := stor.NewTransactor(storage)
	emailSendler := email.NewEmailSender(smtpAuth, cfg.Email)

	authServ := services.NewAuthServ(userRepo, verificationRepo, transactor, emailSendler, log, cfg.Auth)
	readmeServ := services.NewReadmeServ(readmeRepo, userRepo, templateRepo, widgetRepo, transactor, log)
	widgetServ := services.NewWidgetServ(widgetRepo, userRepo, log)
	templateServ := services.NewTemplateServ(templateRepo, userRepo, widgetRepo, transactor, log)
	userServ := services.NewUserServ(userRepo, transactor, log)

	authHandl := handlers.NewAuthHandle(authServ, userServ, validator)
	readmeHandl := handlers.NewReadmeHandl(readmeServ, authServ, validator)
	widgetHandl := handlers.NewWidgetHandl(widgetServ, authServ, validator)
	templateHandl := handlers.NewTemplateHandl(templateServ, authServ, validator)
	userHandl := handlers.NewUserHandl(userServ, validator)

	sheduler := sheduler.NewSheduler(widgetRepo, templateRepo, verificationRepo, cfg.Sheduler, cfg.Search, log)
	sheduler.Start()
	defer func() {
		sheduler.Stop()
		log.Log.Info("sheduler stopped")
	}()
	log.Log.Info("sheduler started")

	routConfig := routs.NewRoutConfig(app.App, userHandl, authHandl, templateHandl, readmeHandl, widgetHandl)
	routConfig.SetupRoutes()

	go func() {
		if err := app.App.Listen(cfg.Server.Host + ":" + cfg.Server.Port); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Log.Info("app shutting down...")
}
