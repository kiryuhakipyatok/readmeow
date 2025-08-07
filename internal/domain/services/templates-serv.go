package services

import (
	"context"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
)

type TemplateServ interface {
	Create(ctx context.Context, oid, title, image, order string, text, links, widgets []string) error
	Update(ctx context.Context, fields map[string]any, id string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*models.Template, error)
	Fetch(ctx context.Context, amount, page uint) ([]models.Template, error)
	Sort(ctx context.Context, amount, page uint, dest, field string) ([]models.Template, error)
}

type templateServ struct {
	TemplateRepo repositories.TemplateRepo
	UserRepo     repositories.UserRepo
	Transactor   storage.Transactor
	Logger       *logger.Logger
}

func NewTemplateServ(tr repositories.TemplateRepo, ur repositories.UserRepo, t storage.Transactor, l *logger.Logger) TemplateServ {
	return &templateServ{
		TemplateRepo: tr,
		UserRepo:     ur,
		Transactor:   t,
		Logger:       l,
	}
}
