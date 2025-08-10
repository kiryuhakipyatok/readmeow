package routs

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
	userGroup.Patch("/password", rc.UserHandl.ChangeUserPassword)
	userGroup.Patch("", rc.UserHandl.Update)
	userGroup.Delete("", rc.UserHandl.Delete)
}

func (rc *RouteConfig) AuthRoutes() {
	authGroup := rc.App.Group("/api/auth")
	authGroup.Post("/register", rc.AuthHandl.Register)
	authGroup.Get("/login", rc.AuthHandl.Login)
	authGroup.Get("/logout", rc.AuthHandl.Logout)
	authGroup.Get("/profile", rc.AuthHandl.Profile)
}

func (rc *RouteConfig) WidgetsRoutes() {
	widgetGroup := rc.App.Group("/api/widgets")

	widgetGroup.Get("/:widget", rc.WidgetHandl.GetWidgetById)
	widgetGroup.Get("", rc.WidgetHandl.FetchWidgets)
	widgetGroup.Get("/sort", rc.WidgetHandl.FetchSortedWidgets)
	widgetGroup.Get("/search", rc.WidgetHandl.SearchWidgets)

	widgetGroup.Patch("/like/:widget", rc.WidgetHandl.Like)
	widgetGroup.Patch("/dislike/:widget", rc.WidgetHandl.Dislike)
}

func (rc *RouteConfig) TemplatesRoutes() {
	templateGroup := rc.App.Group("/api/templates")

	templateGroup.Post("", rc.TemplateHandl.CreateTemplate)

	templateGroup.Delete("/:template", rc.TemplateHandl.DeleteTemplate)

	templateGroup.Get("/:id", rc.TemplateHandl.GetTemplate)
	templateGroup.Get("", rc.TemplateHandl.FetchTemplates)
	templateGroup.Get("/sort", rc.TemplateHandl.SortTemplate)
	templateGroup.Get("/search", rc.TemplateHandl.SearchTemplate)

	templateGroup.Patch("/like/:template", rc.TemplateHandl.Like)
	templateGroup.Patch("/dislike/:template", rc.TemplateHandl.Dislike)
	templateGroup.Patch("", rc.TemplateHandl.UpdateTemplate)
}

func (rc *RouteConfig) ReadmesRoutes() {
	readmeGroup := rc.App.Group("/api/readmes")
	readmeGroup.Post("", rc.ReadmeHandl.CreateReadme)
	readmeGroup.Delete("/:readme", rc.ReadmeHandl.DeleteReadme)
	readmeGroup.Patch("", rc.ReadmeHandl.UpdateReadme)
	readmeGroup.Get("/:id", rc.ReadmeHandl.GetReadmeById)
	readmeGroup.Get("", rc.ReadmeHandl.FetchReadmesByUser)
}
