package bootstrap

import (
	"readmeow/internal/config"
	"readmeow/internal/delivery/handlers"
	"readmeow/internal/delivery/routs"
	"readmeow/internal/domain/repositories"
	"readmeow/internal/domain/services"
	"readmeow/pkg/cache"
	"readmeow/pkg/logger"
	"readmeow/pkg/search"
	"readmeow/pkg/storage"
	"readmeow/pkg/validator"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type BootstrapConfig struct {
	Config       config.Config
	Storage      *storage.Storage
	Cache        *cache.Cache
	SearchClient *search.SearchClient
	Logger       *logger.Logger
	Validator    *validator.Validator
	App          *fiber.App
}

func NewBootstrapConfig(cfg config.Config, app *fiber.App, s *storage.Storage, c *cache.Cache, sc *search.SearchClient, l *logger.Logger, v *validator.Validator) *BootstrapConfig {
	return &BootstrapConfig{
		Config:       cfg,
		Storage:      s,
		Cache:        c,
		SearchClient: sc,
		App:          app,
		Logger:       l,
		Validator:    v,
	}
}

func (bc *BootstrapConfig) Bootstrap() {
	userRepo := repositories.NewUserRepo(bc.Storage)
	widgetRepo := repositories.NewWidgetRepo(bc.Storage, bc.Cache, bc.SearchClient)
	readmeRepo := repositories.NewReadmeStorage(bc.Storage)
	templateRepo := repositories.NewTemplateRepo(bc.Storage, bc.Cache, bc.SearchClient)
	transactor := storage.NewTransactor(bc.Storage)

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		widgetRepo.MustBulk(bc.Config.Search)
	}()
	go func() {
		defer wg.Done()
		templateRepo.MustBulk(bc.Config.Search)
	}()

	authServ := services.NewAuthServ(userRepo, bc.Logger, bc.Config.Auth)
	readmeServ := services.NewReadmeServ(readmeRepo, userRepo, templateRepo, widgetRepo, bc.Logger)
	widgetServ := services.NewWidgetServ(widgetRepo, userRepo, bc.Logger)
	templateServ := services.NewTemplateServ(templateRepo, userRepo, widgetRepo, transactor, bc.Logger)
	userServ := services.NewUserServ(userRepo, bc.Logger)

	authHandl := handlers.NewAuthHandle(authServ, userServ, bc.Validator)
	readmeHandl := handlers.NewReadmeHandl(readmeServ, authServ, bc.Validator)
	widgetHandl := handlers.NewWidgetHandl(widgetServ, authServ, bc.Validator)
	templateHandl := handlers.NewTemplateHandl(templateServ, authServ, bc.Validator)
	userHandl := handlers.NewUserHandl(userServ, bc.Validator)

	routConfig := routs.NewRoutConfig(bc.App, userHandl, authHandl, templateHandl, readmeHandl, widgetHandl)
	routConfig.SetupRoutes()
	wg.Wait()
}
