package routes

import (
	"readmeow/internal/delivery/handlers"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App           *fiber.App
	UserHandl     *handlers.UserHandl
	AuthHandl     *handlers.AuthHandl
	TemplateHandl *handlers.TemplateHandl
	ReadmeHandl   *handlers.ReadmeHandl
	WidgetHandl   *handlers.WidgetHandl
}

func NewRoutConfig(a *fiber.App, uh *handlers.UserHandl, ah *handlers.AuthHandl, th *handlers.TemplateHandl, rh *handlers.ReadmeHandl, wh *handlers.WidgetHandl) *RouteConfig {
	return &RouteConfig{
		App:           a,
		UserHandl:     uh,
		AuthHandl:     ah,
		TemplateHandl: th,
		ReadmeHandl:   rh,
		WidgetHandl:   wh,
	}
}

func (rc *RouteConfig) SetupRoutes() {
	rc.UsersRoutes()
	rc.AuthRoutes()
	rc.ReadmesRoutes()
	rc.TemplatesRoutes()
	rc.WidgetsRoutes()
}

func (rc *RouteConfig) UsersRoutes() {
	userGroup := rc.App.Group("/api/users")
	userGroup.Get("/:user", rc.UserHandl.GetUser)
	userGroup.Patch("", rc.UserHandl.Update)
	userGroup.Patch("/password", rc.UserHandl.ChangeUserPassword)
	userGroup.Delete("", rc.UserHandl.Delete)
}

func (rc *RouteConfig) AuthRoutes() {
	authGroup := rc.App.Group("/api/auth")

	authGroup.Post("/register", rc.AuthHandl.Register)
	authGroup.Post("/verify", rc.AuthHandl.VerifyEmail)
	authGroup.Post("/new-code", rc.AuthHandl.SendNewCode)

	authGroup.Get("/login", rc.AuthHandl.Login)
	authGroup.Get("/logout", rc.AuthHandl.Logout)
	authGroup.Get("/profile", rc.AuthHandl.Profile)

}

func (rc *RouteConfig) WidgetsRoutes() {
	widgetGroup := rc.App.Group("/api/widgets")

	widgetGroup.Get("", rc.WidgetHandl.SearchWidgets)
	widgetGroup.Get("/favorite", rc.WidgetHandl.FetchFavoriteWidgets)
	widgetGroup.Get("/:widget", rc.WidgetHandl.GetWidgetById)

	widgetGroup.Patch("/like/:widget", rc.WidgetHandl.Like)
	widgetGroup.Patch("/dislike/:widget", rc.WidgetHandl.Dislike)
}

func (rc *RouteConfig) TemplatesRoutes() {
	templateGroup := rc.App.Group("/api/templates")

	templateGroup.Post("", rc.TemplateHandl.CreateTemplate)

	templateGroup.Delete("/:template", rc.TemplateHandl.DeleteTemplate)

	templateGroup.Get("", rc.TemplateHandl.SearchTemplate)
	templateGroup.Get("/favorite", rc.TemplateHandl.FetchFavoriteTemplates)
	templateGroup.Get("/:template", rc.TemplateHandl.GetTemplate)

	templateGroup.Patch("", rc.TemplateHandl.UpdateTemplate)
	templateGroup.Patch("/like/:template", rc.TemplateHandl.Like)
	templateGroup.Patch("/dislike/:template", rc.TemplateHandl.Dislike)

}

func (rc *RouteConfig) ReadmesRoutes() {
	readmeGroup := rc.App.Group("/api/readmes")

	readmeGroup.Post("", rc.ReadmeHandl.CreateReadme)

	readmeGroup.Delete("/:readme", rc.ReadmeHandl.DeleteReadme)

	readmeGroup.Patch("", rc.ReadmeHandl.UpdateReadme)

	readmeGroup.Get("", rc.ReadmeHandl.FetchReadmesByUser)
	readmeGroup.Get("/:readme", rc.ReadmeHandl.GetReadmeById)
}
