package bootstrap

// import (
// 	"fmt"
// 	"net/smtp"
// 	"os"
// 	"readmeow/internal/config"
// 	"readmeow/internal/delivery/handlers"
// 	"readmeow/internal/delivery/routs"
// 	"readmeow/internal/domain/repositories"
// 	"readmeow/internal/domain/services"
// 	"readmeow/internal/email"
// 	"readmeow/internal/sheduler"
// 	"readmeow/pkg/cache"
// 	"readmeow/pkg/logger"
// 	"readmeow/pkg/search"
// 	"readmeow/pkg/storage"
// 	"readmeow/pkg/validator"
// 	"sync"

// 	"github.com/gofiber/fiber/v2"
// )

// type BootstrapConfig struct {
// 	Config       config.Config
// 	Storage      *storage.Storage
// 	Cache        *cache.Cache
// 	SearchClient *search.SearchClient
// 	Logger       *logger.Logger
// 	Validator    *validator.Validator
// 	App          *fiber.App
// 	StmpAuth     smtp.Auth
// }

// func NewBootstrapConfig(cfg config.Config, app *fiber.App, s *storage.Storage, c *cache.Cache, sc *search.SearchClient, l *logger.Logger, v *validator.Validator, sa smtp.Auth) *BootstrapConfig {
// 	return &BootstrapConfig{
// 		Config:       cfg,
// 		Storage:      s,
// 		Cache:        c,
// 		SearchClient: sc,
// 		App:          app,
// 		Logger:       l,
// 		Validator:    v,
// 		StmpAuth:     sa,
// 	}
// }

// func (bc *BootstrapConfig) Bootstrap(done <-chan os.Signal) {
// 	userRepo := repositories.NewUserRepo(bc.Storage)
// 	widgetRepo := repositories.NewWidgetRepo(bc.Storage, bc.Cache, bc.SearchClient)
// 	readmeRepo := repositories.NewReadmeStorage(bc.Storage)
// 	templateRepo := repositories.NewTemplateRepo(bc.Storage, bc.Cache, bc.SearchClient)
// 	verificationRepo := repositories.NewVerificationRepo(bc.Storage)
// 	transactor := storage.NewTransactor(bc.Storage)

// 	emailSendler := email.NewEmailSender(bc.StmpAuth, bc.Config.Email)

// 	authServ := services.NewAuthServ(userRepo, verificationRepo, transactor, emailSendler, bc.Logger, bc.Config.Auth)
// 	readmeServ := services.NewReadmeServ(readmeRepo, userRepo, templateRepo, widgetRepo, transactor, bc.Logger)
// 	widgetServ := services.NewWidgetServ(widgetRepo, userRepo, bc.Logger)
// 	templateServ := services.NewTemplateServ(templateRepo, userRepo, widgetRepo, transactor, bc.Logger)
// 	userServ := services.NewUserServ(userRepo, bc.Logger)

// 	authHandl := handlers.NewAuthHandle(authServ, userServ, bc.Validator)
// 	readmeHandl := handlers.NewReadmeHandl(readmeServ, authServ, bc.Validator)
// 	widgetHandl := handlers.NewWidgetHandl(widgetServ, authServ, bc.Validator)
// 	templateHandl := handlers.NewTemplateHandl(templateServ, authServ, bc.Validator)
// 	userHandl := handlers.NewUserHandl(userServ, bc.Validator)

// 	sheduler := sheduler.NewSheduler(widgetRepo, templateRepo, verificationRepo, bc.Config.Sheduler, bc.Config.Search, bc.Logger, done)
// 	var wg sync.WaitGroup
// 	wg.Add(3)
// 	go func() {
// 		defer wg.Done()
// 		sheduler.StartCleanExpiredVerifyCodes()
// 	}()
// 	go func() {
// 		defer wg.Done()
// 		sheduler.StartBulkTemplatesData()
// 	}()
// 	go func() {
// 		defer wg.Done()
// 		sheduler.StartBulkWidgetsData()
// 	}()

// 	fmt.Println("GOOOODA")

// 	routConfig := routs.NewRoutConfig(bc.App, userHandl, authHandl, templateHandl, readmeHandl, widgetHandl)
// 	routConfig.SetupRoutes()

// 	wg.Wait()

// }
