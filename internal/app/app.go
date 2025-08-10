package app

import (
	"context"
	"os"
	"os/signal"
	"readmeow/internal/bootstrap"
	"readmeow/internal/config"
	"readmeow/internal/delivery/server"
	"readmeow/pkg/cache"
	"readmeow/pkg/logger"
	"readmeow/pkg/search"
	"readmeow/pkg/storage"
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
	storage := storage.MustConnect(cfg.Storage)
	log.Log.Info("connected to postgres")
	cache := cache.MustConnect(cfg.Cache)
	log.Log.Info("connected to redis")
	search := search.MustConnect(cfg.Search)
	log.Log.Info("connected to elasticsearch")
	app := server.NewServer(cfg.Server, cfg.Auth)
	log.Log.Info("server created")
	bootstrap := bootstrap.NewBootstrapConfig(*cfg, app.App, storage, cache, search, log, validator)
	bootstrap.Bootstrap()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := app.App.Listen(cfg.Server.Host + ":" + cfg.Server.Port); err != nil {
			panic(err)
		}
	}()
	<-quit
	log.Log.Info("app closing")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.Server.CloseTimeout))
	defer cancel()
	storage.Close()
	log.Log.Info("postgres closed")
	cache.MustClose()
	log.Log.Info("redis closed")
	app.MustClose(ctx)
	log.Log.Info("sercer closed")
	log.Log.Info("app stopped successfully")
}
